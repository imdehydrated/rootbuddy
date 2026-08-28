"""Evaluation harness and baselines for RootBuddy RL agents."""

from __future__ import annotations

import argparse
import time
from collections import Counter, deque
from dataclasses import dataclass, field
from pathlib import Path
from typing import Mapping, Protocol, Sequence

import numpy as np
import torch

from .engine_client import Faction, VecEnvConfig
from .model import CandidatePolicyValueNet, tensors_from_batch
from .train_constants import FACTION_COUNT
from .vec_env import RootBuddyVecEnv, VecEnvArrays, sample_random_actions


class Agent(Protocol):
    def select_actions(self, batch: VecEnvArrays) -> np.ndarray:
        """Return one action index per env row."""


@dataclass
class RandomAgent:
    seed: int | None = None

    def __post_init__(self) -> None:
        self.rng = np.random.default_rng(self.seed)

    def select_actions(self, batch: VecEnvArrays) -> np.ndarray:
        return sample_random_actions(batch, self.rng)


@dataclass(frozen=True)
class GreedyVPAgent:
    def select_actions(self, batch: VecEnvArrays) -> np.ndarray:
        actions = np.full((batch.num_envs,), -1, dtype=np.int64)
        for row, count in enumerate(batch.candidate_counts):
            if count <= 0:
                continue
            rewards = batch.candidate_rewards[row, : int(count)]
            actions[row] = int(np.argmax(rewards))
        return actions


@dataclass
class ModelAgent:
    model: CandidatePolicyValueNet
    deterministic: bool = True
    device: str | torch.device | None = None

    def select_actions(self, batch: VecEnvArrays) -> np.ndarray:
        if np.any(batch.candidate_counts <= 0):
            raise ValueError("model agent cannot select from rows without legal actions")
        was_training = self.model.training
        self.model.eval()
        tensors = tensors_from_batch(batch, device=self.device)
        with torch.no_grad():
            selection = self.model.act(
                tensors["observations"],
                tensors["candidate_actions"],
                tensors["action_mask"],
                tensors["active_factions"],
                deterministic=self.deterministic,
            )
        if was_training:
            self.model.train()
        return selection.actions.detach().cpu().numpy().astype(np.int64)


@dataclass(frozen=True)
class EvaluationConfig:
    env_config: VecEnvConfig
    episodes: int


@dataclass(frozen=True)
class EvaluationDiagnostics:
    terminal_reasons: tuple[tuple[str, int], ...] = ()
    final_active_factions: tuple[tuple[int, int], ...] = ()
    final_phase_steps: tuple[tuple[tuple[int, int], int], ...] = ()
    action_type_counts: tuple[tuple[int, int], ...] = ()
    truncated_recent_action_type_counts: tuple[tuple[int, int], ...] = ()
    per_faction_action_counts: tuple[int, ...] = field(default_factory=lambda: tuple(0 for _ in range(FACTION_COUNT)))
    mean_candidate_count: float = 0.0


@dataclass(frozen=True)
class EvaluationMetrics:
    episodes: int
    transitions: int
    truncated_episodes: int
    mean_reward: float
    mean_game_length: float
    transitions_per_second: float
    per_faction_win_rate: tuple[float, ...]
    per_faction_terminal_vp: tuple[float, ...]
    diagnostics: EvaluationDiagnostics = field(default_factory=EvaluationDiagnostics)


def evaluate_agents(
    config: EvaluationConfig,
    *,
    agents: Mapping[int, Agent],
    default_agent: Agent | None = None,
    env: RootBuddyVecEnv | None = None,
) -> EvaluationMetrics:
    if config.episodes <= 0:
        raise ValueError("episodes must be positive")
    if len(agents) == 0 and default_agent is None:
        raise ValueError("at least one agent or default_agent is required")

    owns_env = env is None
    active_env = env if env is not None else RootBuddyVecEnv(config.env_config)
    fallback = default_agent if default_agent is not None else next(iter(agents.values()))

    episodes = 0
    transitions = 0
    truncated_episodes = 0
    reward_sum = 0.0
    game_lengths: list[int] = []
    winner_counts = [0 for _ in range(FACTION_COUNT)]
    terminal_vp: list[np.ndarray] = []
    terminal_reasons: Counter[str] = Counter()
    final_active_factions: Counter[int] = Counter()
    final_phase_steps: Counter[tuple[int, int]] = Counter()
    action_type_counts: Counter[int] = Counter()
    truncated_recent_action_type_counts: Counter[int] = Counter()
    per_faction_action_counts = [0 for _ in range(FACTION_COUNT)]
    candidate_count_sum = 0
    candidate_count_samples = 0
    recent_actions: dict[int, deque[int]] = {}
    started = time.perf_counter()

    try:
        batch = active_env.reset()
        while episodes < config.episodes:
            done_rows = np.flatnonzero(batch.dones)
            if done_rows.size > 0:
                for row in done_rows:
                    if episodes >= config.episodes:
                        break
                    episodes += 1
                    game_lengths.append(int(batch.steps[row]))
                    terminal_vp.append(batch.victory_points[row].astype(np.float32))
                    terminal_reasons[terminal_reason(batch, int(row), config.env_config.max_steps)] += 1
                    final_active_factions[int(batch.active_factions[row])] += 1
                    final_phase_steps[(int(batch.current_phases[row]), int(batch.current_steps[row]))] += 1
                    recent_action_types = recent_actions_for_row(batch, int(row), recent_actions)
                    recent_actions.pop(int(batch.decisions[int(row)].env_index), None)
                    if bool(batch.truncations[row]):
                        truncated_episodes += 1
                        truncated_recent_action_type_counts.update(recent_action_types)
                        continue
                    winner = int(batch.winners[row])
                    if 0 <= winner < FACTION_COUNT:
                        winner_counts[winner] += 1
                if episodes >= config.episodes:
                    break
                batch = active_env.reset(env_indices=env_indices_for_rows(batch, done_rows))
                continue

            if np.any(batch.candidate_counts <= 0):
                raise ValueError("cannot evaluate env rows without legal actions")
            candidate_count_sum += int(batch.candidate_counts.sum())
            candidate_count_samples += batch.num_envs
            actions = select_actions_by_faction(batch, agents=agents, default_agent=fallback)
            record_selected_actions(
                batch,
                actions,
                recent_actions=recent_actions,
                action_type_counts=action_type_counts,
                per_faction_action_counts=per_faction_action_counts,
            )
            batch = active_env.step(actions.tolist())
            transitions += batch.num_envs
            reward_sum += float(batch.rewards.sum())
    finally:
        if owns_env:
            active_env.close()

    elapsed = max(time.perf_counter() - started, 1e-9)
    return EvaluationMetrics(
        episodes=episodes,
        transitions=transitions,
        truncated_episodes=truncated_episodes,
        mean_reward=reward_sum / max(transitions, 1),
        mean_game_length=float(np.mean(game_lengths)) if game_lengths else 0.0,
        transitions_per_second=transitions / elapsed,
        per_faction_win_rate=tuple(count / episodes for count in winner_counts),
        per_faction_terminal_vp=mean_terminal_vp(terminal_vp),
        diagnostics=EvaluationDiagnostics(
            terminal_reasons=sorted_count_items(terminal_reasons),
            final_active_factions=sorted_count_items(final_active_factions),
            final_phase_steps=sorted_count_items(final_phase_steps),
            action_type_counts=sorted_count_items(action_type_counts),
            truncated_recent_action_type_counts=sorted_count_items(truncated_recent_action_type_counts),
            per_faction_action_counts=tuple(per_faction_action_counts),
            mean_candidate_count=candidate_count_sum / max(candidate_count_samples, 1),
        ),
    )


def env_indices_for_rows(batch: VecEnvArrays, rows: np.ndarray) -> list[int]:
    return [int(batch.decisions[int(row)].env_index) for row in rows]


def terminal_reason(batch: VecEnvArrays, row: int, max_steps: int) -> str:
    if not bool(batch.truncations[row]):
        return "win"
    if int(batch.steps[row]) < max_steps:
        return "no_legal"
    return "max_steps"


def recent_actions_for_row(
    batch: VecEnvArrays,
    row: int,
    recent_actions: Mapping[int, deque[int]],
) -> tuple[int, ...]:
    env_index = int(batch.decisions[row].env_index)
    return tuple(recent_actions.get(env_index, ()))


def record_selected_actions(
    batch: VecEnvArrays,
    actions: np.ndarray,
    *,
    recent_actions: dict[int, deque[int]],
    action_type_counts: Counter[int],
    per_faction_action_counts: list[int],
) -> None:
    for row, action_index in enumerate(actions):
        index = int(action_index)
        if index < 0 or index >= len(batch.decisions[row].legal_action_types):
            continue
        action_type = int(batch.decisions[row].legal_action_types[index])
        action_type_counts[action_type] += 1
        faction = int(batch.active_factions[row])
        if 0 <= faction < len(per_faction_action_counts):
            per_faction_action_counts[faction] += 1
        env_index = int(batch.decisions[row].env_index)
        recent_actions.setdefault(env_index, deque(maxlen=64)).append(action_type)


def sorted_count_items(counter: Mapping) -> tuple:
    return tuple(sorted(counter.items(), key=lambda item: item[0]))


def select_actions_by_faction(
    batch: VecEnvArrays,
    *,
    agents: Mapping[int, Agent],
    default_agent: Agent,
) -> np.ndarray:
    actions = np.full((batch.num_envs,), -1, dtype=np.int64)
    proposals: dict[int, np.ndarray] = {}
    for faction in np.unique(batch.active_factions):
        faction_id = int(faction)
        agent = agents.get(faction_id, default_agent)
        proposal_key = id(agent)
        if proposal_key not in proposals:
            proposal = agent.select_actions(batch)
            if proposal.shape != (batch.num_envs,):
                raise ValueError("agent returned wrong action shape")
            proposals[proposal_key] = proposal
        rows = batch.active_factions == faction_id
        actions[rows] = proposals[proposal_key][rows]
    if np.any(actions < 0):
        raise ValueError("agent selection left invalid action rows")
    return actions


def mean_terminal_vp(values: list[np.ndarray]) -> tuple[float, ...]:
    if len(values) == 0:
        return tuple(0.0 for _ in range(FACTION_COUNT))
    stacked = np.stack(values, axis=0)
    return tuple(float(value) for value in stacked.mean(axis=0))


def load_model_from_checkpoint(
    checkpoint_path: str | Path,
    *,
    observation_dim: int,
    action_dim: int,
    device: str | torch.device | None = None,
) -> CandidatePolicyValueNet:
    checkpoint = torch.load(checkpoint_path, map_location=device, weights_only=False)
    config = dict(checkpoint.get("config") or {})
    model = CandidatePolicyValueNet(
        observation_dim=observation_dim,
        action_dim=action_dim,
        torso_hidden_dims=tuple(config.get("torso_hidden_dims", (256, 256))),
        head_hidden_dim=int(config.get("head_hidden_dim", 128)),
    )
    model.load_state_dict(checkpoint["model_state_dict"])
    if device is not None:
        model.to(device)
    model.eval()
    return model


def baseline_agent(name: str, *, seed: int | None = None) -> Agent:
    if name == "random":
        return RandomAgent(seed=seed)
    if name == "greedy-vp":
        return GreedyVPAgent()
    raise ValueError(f"unknown baseline {name!r}")


def two_player_eval_config(args: argparse.Namespace) -> EvaluationConfig:
    return EvaluationConfig(
        env_config=VecEnvConfig(
            num_envs=args.num_envs,
            base_seed=args.base_seed,
            max_steps=args.max_steps,
            factions=[Faction.MARQUISE, Faction.EYRIE],
            player_faction=Faction.MARQUISE,
            track_all_hands=not args.partial_observability,
            step_penalty=args.step_penalty,
            truncation_penalty=args.truncation_penalty,
        ),
        episodes=args.episodes,
    )


def parse_factions(value: str) -> tuple[int, ...]:
    if value.strip() == "":
        return ()
    return tuple(int(part.strip()) for part in value.split(","))


def parse_args(argv: Sequence[str] | None = None) -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="Evaluate RootBuddy RL agents.")
    parser.add_argument("--episodes", type=int, default=8)
    parser.add_argument("--num-envs", type=int, default=2)
    parser.add_argument("--base-seed", type=int, default=1707)
    parser.add_argument("--max-steps", type=int, default=1024)
    parser.add_argument("--baseline", choices=["random", "greedy-vp"], default="random")
    parser.add_argument("--model-checkpoint", default=None)
    parser.add_argument("--model-factions", default="0")
    parser.add_argument("--stochastic-model", action="store_true")
    parser.add_argument("--device", default=None)
    parser.add_argument("--seed", type=int, default=0)
    parser.add_argument("--step-penalty", type=float, default=0.01)
    parser.add_argument("--truncation-penalty", type=float, default=10.0)
    parser.add_argument(
        "--partial-observability",
        action="store_true",
        help="request hidden opponent hands; currently overridden by the Go RL env until hidden-card action generation is supported",
    )
    parser.add_argument("--track-all-hands", action="store_true", help=argparse.SUPPRESS)
    return parser.parse_args(argv)


def main(argv: Sequence[str] | None = None) -> None:
    args = parse_args(argv)
    config = two_player_eval_config(args)
    baseline = baseline_agent(args.baseline, seed=args.seed)
    agents: dict[int, Agent] = {}

    with RootBuddyVecEnv(config.env_config) as env:
        initial_batch = env.reset()
        if args.model_checkpoint is not None:
            model = load_model_from_checkpoint(
                args.model_checkpoint,
                observation_dim=initial_batch.observation_length,
                action_dim=initial_batch.action_length,
                device=args.device,
            )
            model_agent = ModelAgent(
                model=model,
                deterministic=not args.stochastic_model,
                device=args.device,
            )
            for faction in parse_factions(args.model_factions):
                agents[faction] = model_agent
        metrics = evaluate_agents(config, agents=agents, default_agent=baseline, env=env)

    print(format_metrics(metrics))


def format_metrics(metrics: EvaluationMetrics) -> str:
    return (
        "episodes={episodes} transitions={transitions} truncated={truncated} "
        "reward={reward:.4f} game_length={length:.2f} tps={tps:.1f} "
        "win_rate={wins} terminal_vp={vp} diagnostics={diagnostics}"
    ).format(
        episodes=metrics.episodes,
        transitions=metrics.transitions,
        truncated=metrics.truncated_episodes,
        reward=metrics.mean_reward,
        length=metrics.mean_game_length,
        tps=metrics.transitions_per_second,
        wins=tuple(round(value, 4) for value in metrics.per_faction_win_rate),
        vp=tuple(round(value, 2) for value in metrics.per_faction_terminal_vp),
        diagnostics=format_diagnostics(metrics.diagnostics),
    )


def format_diagnostics(diagnostics: EvaluationDiagnostics) -> str:
    return {
        "terminal_reasons": dict(diagnostics.terminal_reasons),
        "final_active_factions": dict(diagnostics.final_active_factions),
        "final_phase_steps": {f"{phase}:{step}": count for (phase, step), count in diagnostics.final_phase_steps},
        "action_type_counts": dict(diagnostics.action_type_counts),
        "truncated_recent_action_type_counts": dict(diagnostics.truncated_recent_action_type_counts),
        "per_faction_action_counts": diagnostics.per_faction_action_counts,
        "mean_candidate_count": round(diagnostics.mean_candidate_count, 2),
    }


if __name__ == "__main__":
    main()
