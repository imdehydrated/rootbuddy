"""Training loop for RootBuddy PPO self-play."""

from __future__ import annotations

import argparse
import time
from dataclasses import asdict, dataclass, replace
from pathlib import Path
from typing import Sequence

import torch

from .engine_client import Faction, VecEnvConfig
from .model import CandidatePolicyValueNet
from .ppo import PPOConfig, PPOUpdateStats, ppo_update
from .rollout import RolloutConfig, RolloutStats, collect_rollout
from .vec_env import RootBuddyVecEnv, VecEnvArrays

FACTION_COUNT = 4


@dataclass(frozen=True)
class TrainConfig:
    env_config: VecEnvConfig
    updates: int
    rollout_steps: int
    learning_rate: float = 3e-4
    gamma: float = 0.99
    gae_lambda: float = 0.95
    ppo_epochs: int = 4
    minibatch_size: int = 256
    clip_range: float = 0.2
    value_coef: float = 0.5
    entropy_coef: float = 0.01
    max_grad_norm: float = 0.5
    checkpoint_dir: str | Path = "rl/checkpoints"
    checkpoint_every: int = 10
    log_dir: str | Path | None = None
    device: str | torch.device | None = None
    torso_hidden_dims: tuple[int, ...] = (256, 256)
    head_hidden_dim: int = 128
    seed: int = 0
    stop_on_done: bool = True


@dataclass(frozen=True)
class TrainMetrics:
    update: int
    rollout_steps: int
    transitions: int
    completed_episodes: int
    mean_reward: float
    mean_game_length: float
    per_faction_win_rate: tuple[float, ...]
    per_faction_terminal_vp: tuple[float, ...]
    transitions_per_second: float
    loss: float
    policy_loss: float
    value_loss: float
    entropy: float
    approx_kl: float
    clip_fraction: float
    checkpoint_path: str | None = None


@dataclass(frozen=True)
class TrainResult:
    model: CandidatePolicyValueNet
    optimizer: torch.optim.Optimizer
    metrics: tuple[TrainMetrics, ...]


def train(
    config: TrainConfig,
    *,
    env: RootBuddyVecEnv | None = None,
    model: CandidatePolicyValueNet | None = None,
    optimizer: torch.optim.Optimizer | None = None,
) -> TrainResult:
    if config.updates <= 0:
        raise ValueError("updates must be positive")
    if config.rollout_steps <= 0:
        raise ValueError("rollout_steps must be positive")
    if config.seed != 0:
        torch.manual_seed(config.seed)

    device = torch.device(config.device) if config.device is not None else torch.device("cpu")
    owns_env = env is None
    active_env = env if env is not None else RootBuddyVecEnv(config.env_config)
    writer = create_summary_writer(config.log_dir)

    try:
        initial_batch = active_env.reset()
        if model is None:
            model = build_model_from_batch(initial_batch, config)
        model.to(device)
        if optimizer is None:
            optimizer = torch.optim.Adam(model.parameters(), lr=config.learning_rate)

        metrics: list[TrainMetrics] = []
        for update in range(1, config.updates + 1):
            started = time.perf_counter()
            rollout = collect_rollout(
                active_env,
                model,
                RolloutConfig(
                    steps=config.rollout_steps,
                    gamma=config.gamma,
                    gae_lambda=config.gae_lambda,
                    device=device,
                    stop_on_done=config.stop_on_done,
                ),
            )
            ppo_stats = ppo_update(
                model,
                optimizer,
                rollout.batch.to(device),
                PPOConfig(
                    clip_range=config.clip_range,
                    value_coef=config.value_coef,
                    entropy_coef=config.entropy_coef,
                    max_grad_norm=config.max_grad_norm,
                    epochs=config.ppo_epochs,
                    minibatch_size=config.minibatch_size,
                ),
            )
            elapsed = max(time.perf_counter() - started, 1e-9)
            metric = make_train_metrics(
                update=update,
                rollout_stats=rollout.stats,
                last_env_batch=rollout.last_env_batch,
                ppo_stats=ppo_stats,
                elapsed=elapsed,
            )
            checkpoint_path = maybe_save_checkpoint(
                config,
                update=update,
                model=model,
                optimizer=optimizer,
                metrics=[*metrics, metric],
                ppo_stats=ppo_stats,
            )
            if checkpoint_path is not None:
                metric = replace(metric, checkpoint_path=str(checkpoint_path))
            metrics.append(metric)
            log_metrics(writer, metric)
        return TrainResult(model=model, optimizer=optimizer, metrics=tuple(metrics))
    finally:
        if writer is not None:
            writer.close()
        if owns_env:
            active_env.close()


def build_model_from_batch(batch: VecEnvArrays, config: TrainConfig) -> CandidatePolicyValueNet:
    return CandidatePolicyValueNet(
        observation_dim=batch.observation_length,
        action_dim=batch.action_length,
        torso_hidden_dims=config.torso_hidden_dims,
        head_hidden_dim=config.head_hidden_dim,
    )


def make_train_metrics(
    *,
    update: int,
    rollout_stats: RolloutStats,
    last_env_batch: VecEnvArrays,
    ppo_stats: PPOUpdateStats,
    elapsed: float,
    checkpoint_path: Path | None = None,
) -> TrainMetrics:
    outcomes = terminal_outcomes(last_env_batch)
    return TrainMetrics(
        update=update,
        rollout_steps=rollout_stats.steps,
        transitions=rollout_stats.transitions,
        completed_episodes=rollout_stats.completed_episodes,
        mean_reward=rollout_stats.mean_reward,
        mean_game_length=outcomes["mean_game_length"],
        per_faction_win_rate=outcomes["per_faction_win_rate"],
        per_faction_terminal_vp=outcomes["per_faction_terminal_vp"],
        transitions_per_second=rollout_stats.transitions / elapsed,
        loss=ppo_stats.loss,
        policy_loss=ppo_stats.policy_loss,
        value_loss=ppo_stats.value_loss,
        entropy=ppo_stats.entropy,
        approx_kl=ppo_stats.approx_kl,
        clip_fraction=ppo_stats.clip_fraction,
        checkpoint_path=str(checkpoint_path) if checkpoint_path is not None else None,
    )


def terminal_outcomes(batch: VecEnvArrays) -> dict[str, float | tuple[float, ...]]:
    done_rows = batch.dones.astype(bool)
    completed = int(done_rows.sum())
    if completed == 0:
        return {
            "mean_game_length": 0.0,
            "per_faction_win_rate": tuple(0.0 for _ in range(FACTION_COUNT)),
            "per_faction_terminal_vp": tuple(0.0 for _ in range(FACTION_COUNT)),
        }

    winner_counts = [0 for _ in range(FACTION_COUNT)]
    decided_rows = done_rows & ~batch.truncations.astype(bool)
    for winner in batch.winners[decided_rows]:
        if 0 <= int(winner) < FACTION_COUNT:
            winner_counts[int(winner)] += 1

    terminal_vp = batch.victory_points[done_rows]
    return {
        "mean_game_length": float(batch.steps[done_rows].mean()),
        "per_faction_win_rate": tuple(count / completed for count in winner_counts),
        "per_faction_terminal_vp": tuple(float(value) for value in terminal_vp.mean(axis=0)),
    }


def maybe_save_checkpoint(
    config: TrainConfig,
    *,
    update: int,
    model: CandidatePolicyValueNet,
    optimizer: torch.optim.Optimizer,
    metrics: Sequence[TrainMetrics],
    ppo_stats: PPOUpdateStats,
) -> Path | None:
    if config.checkpoint_every <= 0:
        return None
    if update % config.checkpoint_every != 0 and update != config.updates:
        return None

    checkpoint_dir = Path(config.checkpoint_dir)
    checkpoint_dir.mkdir(parents=True, exist_ok=True)
    path = checkpoint_dir / f"checkpoint_{update:06d}.pt"
    torch.save(
        {
            "update": update,
            "model_state_dict": model.state_dict(),
            "optimizer_state_dict": optimizer.state_dict(),
            "config": config_to_dict(config),
            "metrics": [asdict(metric) for metric in metrics],
            "latest_ppo_stats": asdict(ppo_stats),
        },
        path,
    )
    return path


def config_to_dict(config: TrainConfig) -> dict[str, object]:
    data = asdict(config)
    env_config = config.env_config.to_wire()
    data["env_config"] = env_config
    data["checkpoint_dir"] = str(config.checkpoint_dir)
    data["log_dir"] = str(config.log_dir) if config.log_dir is not None else None
    data["device"] = str(config.device) if config.device is not None else None
    return data


def create_summary_writer(log_dir: str | Path | None):
    if log_dir is None:
        return None
    from torch.utils.tensorboard import SummaryWriter

    return SummaryWriter(log_dir=str(log_dir))


def log_metrics(writer, metrics: TrainMetrics) -> None:
    if writer is None:
        return
    prefix = "train"
    writer.add_scalar(f"{prefix}/mean_reward", metrics.mean_reward, metrics.update)
    writer.add_scalar(f"{prefix}/mean_game_length", metrics.mean_game_length, metrics.update)
    writer.add_scalar(f"{prefix}/completed_episodes", metrics.completed_episodes, metrics.update)
    writer.add_scalar(f"{prefix}/transitions_per_second", metrics.transitions_per_second, metrics.update)
    writer.add_scalar(f"{prefix}/loss", metrics.loss, metrics.update)
    writer.add_scalar(f"{prefix}/policy_loss", metrics.policy_loss, metrics.update)
    writer.add_scalar(f"{prefix}/value_loss", metrics.value_loss, metrics.update)
    writer.add_scalar(f"{prefix}/entropy", metrics.entropy, metrics.update)
    writer.add_scalar(f"{prefix}/approx_kl", metrics.approx_kl, metrics.update)
    writer.add_scalar(f"{prefix}/clip_fraction", metrics.clip_fraction, metrics.update)
    for faction, value in enumerate(metrics.per_faction_win_rate):
        writer.add_scalar(f"{prefix}/faction_{faction}_win_rate", value, metrics.update)
    for faction, value in enumerate(metrics.per_faction_terminal_vp):
        writer.add_scalar(f"{prefix}/faction_{faction}_terminal_vp", value, metrics.update)


def two_player_train_config(args: argparse.Namespace) -> TrainConfig:
    return TrainConfig(
        env_config=VecEnvConfig(
            num_envs=args.num_envs,
            base_seed=args.base_seed,
            max_steps=args.max_steps,
            factions=[Faction.MARQUISE, Faction.EYRIE],
            player_faction=Faction.MARQUISE,
            track_all_hands=args.track_all_hands,
        ),
        updates=args.updates,
        rollout_steps=args.rollout_steps,
        learning_rate=args.learning_rate,
        ppo_epochs=args.ppo_epochs,
        minibatch_size=args.minibatch_size,
        checkpoint_dir=args.checkpoint_dir,
        checkpoint_every=args.checkpoint_every,
        log_dir=args.log_dir,
        seed=args.seed,
    )


def parse_args(argv: Sequence[str] | None = None) -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="Train RootBuddy PPO agents.")
    parser.add_argument("--updates", type=int, default=10)
    parser.add_argument("--num-envs", type=int, default=4)
    parser.add_argument("--base-seed", type=int, default=707)
    parser.add_argument("--max-steps", type=int, default=256)
    parser.add_argument("--rollout-steps", type=int, default=64)
    parser.add_argument("--learning-rate", type=float, default=3e-4)
    parser.add_argument("--ppo-epochs", type=int, default=4)
    parser.add_argument("--minibatch-size", type=int, default=256)
    parser.add_argument("--checkpoint-dir", default="rl/checkpoints")
    parser.add_argument("--checkpoint-every", type=int, default=10)
    parser.add_argument("--log-dir", default=None)
    parser.add_argument("--seed", type=int, default=0)
    parser.add_argument("--track-all-hands", action="store_true")
    return parser.parse_args(argv)


def main(argv: Sequence[str] | None = None) -> None:
    config = two_player_train_config(parse_args(argv))
    result = train(config)
    for metrics in result.metrics:
        print(
            "update={update} transitions={transitions} reward={reward:.4f} "
            "loss={loss:.4f} entropy={entropy:.4f} tps={tps:.1f}".format(
                update=metrics.update,
                transitions=metrics.transitions,
                reward=metrics.mean_reward,
                loss=metrics.loss,
                entropy=metrics.entropy,
                tps=metrics.transitions_per_second,
            )
        )


if __name__ == "__main__":
    main()
