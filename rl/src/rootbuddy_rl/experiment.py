"""Train/evaluate experiment runner for reward and PPO tuning."""

from __future__ import annotations

import argparse
from dataclasses import dataclass, replace
from pathlib import Path
from typing import Callable, Mapping, Sequence

from .engine_client import Faction, VecEnvConfig
from .evaluate import EvaluationConfig, EvaluationMetrics, ModelAgent, baseline_agent, evaluate_agents, format_metrics
from .train import TrainConfig, TrainMetrics, TrainResult, train
from .train_constants import FACTION_COUNT


@dataclass(frozen=True)
class ExperimentConfig:
    name: str
    train_config: TrainConfig
    eval_episodes: int
    eval_num_envs: int
    eval_base_seed: int
    baselines: tuple[str, ...] = ("random", "greedy-vp")
    model_factions: tuple[int, ...] = (int(Faction.MARQUISE),)
    max_truncation_rate: float = 0.25
    min_entropy: float = 0.05
    max_approx_kl: float = 0.25
    max_win_rate_gap: float = 0.75


@dataclass(frozen=True)
class ExperimentDiagnostics:
    warnings: tuple[str, ...]
    final_entropy: float
    max_approx_kl: float
    max_truncation_rate: float
    max_win_rate_gap: float


@dataclass(frozen=True)
class ExperimentResult:
    config: ExperimentConfig
    train_result: TrainResult
    evaluations: Mapping[str, EvaluationMetrics]
    diagnostics: ExperimentDiagnostics


TrainFn = Callable[[TrainConfig], TrainResult]
EvaluateFn = Callable[..., EvaluationMetrics]


def run_experiment(
    config: ExperimentConfig,
    *,
    train_fn: TrainFn = train,
    evaluate_fn: EvaluateFn = evaluate_agents,
) -> ExperimentResult:
    train_result = train_fn(config.train_config)
    if len(train_result.metrics) == 0:
        raise ValueError("experiment training produced no metrics")

    model_agent = ModelAgent(train_result.model, deterministic=True, device=config.train_config.device)
    eval_config = evaluation_config_from_train(config)
    evaluations: dict[str, EvaluationMetrics] = {}
    for index, baseline in enumerate(config.baselines):
        evaluations[baseline] = evaluate_fn(
            eval_config,
            agents={faction: model_agent for faction in config.model_factions},
            default_agent=baseline_agent(baseline, seed=config.eval_base_seed + index),
        )

    diagnostics = diagnose_experiment(
        train_result.metrics,
        evaluations,
        max_truncation_rate=config.max_truncation_rate,
        min_entropy=config.min_entropy,
        max_approx_kl=config.max_approx_kl,
        max_win_rate_gap=config.max_win_rate_gap,
    )
    return ExperimentResult(
        config=config,
        train_result=train_result,
        evaluations=evaluations,
        diagnostics=diagnostics,
    )


def evaluation_config_from_train(config: ExperimentConfig) -> EvaluationConfig:
    train_env = config.train_config.env_config
    return EvaluationConfig(
        env_config=VecEnvConfig(
            num_envs=config.eval_num_envs,
            base_seed=config.eval_base_seed,
            max_steps=train_env.max_steps,
            factions=train_env.factions,
            player_faction=train_env.player_faction,
            map_id=train_env.map_id,
            track_all_hands=train_env.track_all_hands,
            terminal_win_bonus=train_env.terminal_win_bonus,
        ),
        episodes=config.eval_episodes,
    )


def diagnose_experiment(
    train_metrics: Sequence[TrainMetrics],
    evaluations: Mapping[str, EvaluationMetrics],
    *,
    max_truncation_rate: float,
    min_entropy: float,
    max_approx_kl: float,
    max_win_rate_gap: float,
) -> ExperimentDiagnostics:
    if len(train_metrics) == 0:
        raise ValueError("cannot diagnose an experiment without training metrics")

    final_entropy = train_metrics[-1].entropy
    observed_max_kl = max(abs(metric.approx_kl) for metric in train_metrics)
    observed_max_truncation = max(
        (metrics.truncated_episodes / max(metrics.episodes, 1) for metrics in evaluations.values()),
        default=0.0,
    )
    observed_max_gap = max(
        (win_rate_gap(metrics.per_faction_win_rate) for metrics in evaluations.values()),
        default=0.0,
    )

    warnings = []
    if final_entropy < min_entropy:
        warnings.append("entropy is below the tuning floor; exploration may be collapsing")
    if observed_max_kl > max_approx_kl:
        warnings.append("approx_kl exceeded the tuning ceiling; reduce learning rate or PPO epochs")
    if observed_max_truncation > max_truncation_rate:
        warnings.append("too many eval games truncated; increase max steps or inspect stalled play")
    if observed_max_gap > max_win_rate_gap:
        warnings.append("faction win rates are highly imbalanced; inspect reward shaping and matchups")

    return ExperimentDiagnostics(
        warnings=tuple(warnings),
        final_entropy=final_entropy,
        max_approx_kl=observed_max_kl,
        max_truncation_rate=observed_max_truncation,
        max_win_rate_gap=observed_max_gap,
    )


def win_rate_gap(win_rates: Sequence[float]) -> float:
    if len(win_rates) == 0:
        return 0.0
    live_rates = tuple(float(value) for value in win_rates[:FACTION_COUNT])
    return max(live_rates) - min(live_rates)


def experiment_preset(name: str, *, output_dir: str | Path = "rl/runs") -> ExperimentConfig:
    root = Path(output_dir) / name
    base_env = VecEnvConfig(
        num_envs=2,
        base_seed=707,
        max_steps=64,
        factions=[Faction.MARQUISE, Faction.EYRIE],
        player_faction=Faction.MARQUISE,
        terminal_win_bonus=30.0,
    )
    if name == "smoke":
        return ExperimentConfig(
            name=name,
            train_config=TrainConfig(
                env_config=base_env,
                updates=2,
                rollout_steps=4,
                ppo_epochs=1,
                minibatch_size=4,
                checkpoint_dir=root / "checkpoints",
                checkpoint_every=1,
                log_dir=root / "tb",
                seed=707,
                league_opponent_fraction=0.25,
                league_max_snapshots=2,
                league_snapshot_dir=root / "checkpoints",
                league_seed=707,
            ),
            eval_episodes=2,
            eval_num_envs=1,
            eval_base_seed=1707,
        )
    if name == "two-player-fast":
        return ExperimentConfig(
            name=name,
            train_config=TrainConfig(
                env_config=replace(base_env, num_envs=4, max_steps=256),
                updates=25,
                rollout_steps=32,
                ppo_epochs=2,
                minibatch_size=128,
                learning_rate=3e-4,
                entropy_coef=0.02,
                checkpoint_dir=root / "checkpoints",
                checkpoint_every=5,
                log_dir=root / "tb",
                seed=808,
                league_opponent_fraction=0.25,
                league_max_snapshots=4,
                league_snapshot_dir=root / "checkpoints",
                league_seed=808,
            ),
            eval_episodes=8,
            eval_num_envs=2,
            eval_base_seed=1808,
        )
    if name == "two-player-stable":
        return ExperimentConfig(
            name=name,
            train_config=TrainConfig(
                env_config=replace(base_env, num_envs=8, max_steps=512),
                updates=100,
                rollout_steps=64,
                ppo_epochs=4,
                minibatch_size=256,
                learning_rate=2.5e-4,
                entropy_coef=0.01,
                checkpoint_dir=root / "checkpoints",
                checkpoint_every=10,
                log_dir=root / "tb",
                seed=909,
                league_opponent_fraction=0.35,
                league_max_snapshots=8,
                league_snapshot_dir=root / "checkpoints",
                league_seed=909,
            ),
            eval_episodes=24,
            eval_num_envs=4,
            eval_base_seed=1909,
            max_truncation_rate=0.20,
            max_win_rate_gap=0.65,
        )
    raise ValueError(f"unknown experiment preset {name!r}")


def apply_overrides(config: ExperimentConfig, args: argparse.Namespace) -> ExperimentConfig:
    train_config = config.train_config
    if args.updates is not None:
        train_config = replace(train_config, updates=args.updates)
    if args.rollout_steps is not None:
        train_config = replace(train_config, rollout_steps=args.rollout_steps)
    if args.learning_rate is not None:
        train_config = replace(train_config, learning_rate=args.learning_rate)
    if args.entropy_coef is not None:
        train_config = replace(train_config, entropy_coef=args.entropy_coef)
    if args.terminal_win_bonus is not None:
        train_config = replace(
            train_config,
            env_config=replace(train_config.env_config, terminal_win_bonus=args.terminal_win_bonus),
        )
    if args.league_opponent_fraction is not None:
        train_config = replace(train_config, league_opponent_fraction=args.league_opponent_fraction)
    if args.device is not None:
        train_config = replace(train_config, device=args.device)
    if args.output_dir is not None:
        root = Path(args.output_dir) / config.name
        train_config = replace(
            train_config,
            checkpoint_dir=root / "checkpoints",
            log_dir=root / "tb",
            league_snapshot_dir=root / "checkpoints",
        )
    return replace(
        config,
        train_config=train_config,
        eval_episodes=args.eval_episodes if args.eval_episodes is not None else config.eval_episodes,
        eval_num_envs=args.eval_num_envs if args.eval_num_envs is not None else config.eval_num_envs,
        baselines=tuple(args.baselines.split(",")) if args.baselines is not None else config.baselines,
    )


def parse_args(argv: Sequence[str] | None = None) -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="Run RootBuddy train/eval tuning experiments.")
    parser.add_argument("--preset", choices=["smoke", "two-player-fast", "two-player-stable"], default="smoke")
    parser.add_argument("--output-dir", default=None)
    parser.add_argument("--updates", type=int, default=None)
    parser.add_argument("--rollout-steps", type=int, default=None)
    parser.add_argument("--learning-rate", type=float, default=None)
    parser.add_argument("--entropy-coef", type=float, default=None)
    parser.add_argument("--terminal-win-bonus", type=float, default=None)
    parser.add_argument("--league-opponent-fraction", type=float, default=None)
    parser.add_argument("--eval-episodes", type=int, default=None)
    parser.add_argument("--eval-num-envs", type=int, default=None)
    parser.add_argument("--baselines", default=None)
    parser.add_argument("--device", default=None)
    return parser.parse_args(argv)


def main(argv: Sequence[str] | None = None) -> None:
    args = parse_args(argv)
    config = apply_overrides(experiment_preset(args.preset, output_dir=args.output_dir or "rl/runs"), args)
    result = run_experiment(config)
    print(format_experiment_result(result))


def format_experiment_result(result: ExperimentResult) -> str:
    final_train = result.train_result.metrics[-1]
    lines = [
        f"preset={result.config.name}",
        (
            "train "
            f"updates={len(result.train_result.metrics)} "
            f"reward={final_train.mean_reward:.4f} "
            f"entropy={final_train.entropy:.4f} "
            f"kl={final_train.approx_kl:.4f} "
            f"league_pool={final_train.league_pool_size}"
        ),
    ]
    for name, metrics in result.evaluations.items():
        lines.append(f"eval[{name}] {format_metrics(metrics)}")
    if result.diagnostics.warnings:
        lines.append("warnings=" + "; ".join(result.diagnostics.warnings))
    else:
        lines.append("warnings=none")
    return "\n".join(lines)


if __name__ == "__main__":
    main()
