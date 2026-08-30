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

ACTION_BUILD = 3
ACTION_RECRUIT = 4
ACTION_ADD_TO_DECREE = 7
ACTION_TURMOIL = 18
ACTION_SCORE_ROOSTS = 23
ACTION_EYRIE_SETUP = 33
ACTION_EYRIE_NEW_ROOST = 37
FACTION_MARQUISE = int(Faction.MARQUISE)
FACTION_EYRIE = int(Faction.EYRIE)


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
    mean_terminal_round_number: float = 0.0
    eyrie_action_type_counts: tuple[tuple[int, int], ...] = ()
    eyrie_turmoil_count: int = 0
    eyrie_turmoil_vp_lost: int = 0
    eyrie_turmoil_by_decree_column: tuple[tuple[int, int], ...] = ()
    eyrie_score_roosts_count: int = 0
    eyrie_score_roosts_vp: int = 0
    eyrie_roost_builds: int = 0
    eyrie_roost_losses: int = 0
    eyrie_roost_losses_on_marquise_turn: int = 0
    eyrie_cards_added_by_decree_column: tuple[int, int, int, int] = (0, 0, 0, 0)
    mean_eyrie_roosts: float = 0.0
    mean_terminal_eyrie_roosts: float = 0.0
    mean_terminal_eyrie_vp: float = 0.0
    mean_terminal_eyrie_decree_counts: tuple[float, float, float, float] = (0.0, 0.0, 0.0, 0.0)


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
    terminal_round_numbers: list[int] = []
    terminal_eyrie_roosts: list[int] = []
    terminal_eyrie_vp: list[int] = []
    terminal_eyrie_decree_counts: list[np.ndarray] = []
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
    eyrie_roost_sum = 0
    eyrie_roost_samples = 0
    eyrie_action_type_counts: Counter[int] = Counter()
    eyrie_turmoil_count = 0
    eyrie_turmoil_vp_lost = 0
    eyrie_turmoil_by_decree_column: Counter[int] = Counter()
    eyrie_score_roosts_count = 0
    eyrie_score_roosts_vp = 0
    eyrie_roost_builds = 0
    eyrie_roost_losses = 0
    eyrie_roost_losses_on_marquise_turn = 0
    eyrie_cards_added_by_decree_column = np.zeros((4,), dtype=np.int64)
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
                    terminal_round_numbers.append(int(batch.round_numbers[row]))
                    terminal_eyrie_roosts.append(int(batch.eyrie_roosts[row]))
                    terminal_eyrie_vp.append(int(batch.victory_points[row, FACTION_EYRIE]))
                    terminal_eyrie_decree_counts.append(batch.eyrie_decree_counts[row].astype(np.float32))
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
            eyrie_roost_sum += int(batch.eyrie_roosts.sum())
            eyrie_roost_samples += batch.num_envs
            actions = select_actions_by_faction(batch, agents=agents, default_agent=fallback)
            record_selected_actions(
                batch,
                actions,
                recent_actions=recent_actions,
                action_type_counts=action_type_counts,
                per_faction_action_counts=per_faction_action_counts,
            )
            previous = batch
            batch = active_env.step(actions.tolist())
            transition_diagnostics = eyrie_transition_diagnostics(previous, batch, actions)
            eyrie_action_type_counts.update(transition_diagnostics.action_type_counts)
            eyrie_turmoil_count += transition_diagnostics.turmoil_count
            eyrie_turmoil_vp_lost += transition_diagnostics.turmoil_vp_lost
            eyrie_turmoil_by_decree_column.update(transition_diagnostics.turmoil_by_decree_column)
            eyrie_score_roosts_count += transition_diagnostics.score_roosts_count
            eyrie_score_roosts_vp += transition_diagnostics.score_roosts_vp
            eyrie_roost_builds += transition_diagnostics.roost_builds
            eyrie_roost_losses += transition_diagnostics.roost_losses
            eyrie_roost_losses_on_marquise_turn += transition_diagnostics.roost_losses_on_marquise_turn
            eyrie_cards_added_by_decree_column += transition_diagnostics.cards_added_by_decree_column
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
            mean_terminal_round_number=float(np.mean(terminal_round_numbers)) if terminal_round_numbers else 0.0,
            eyrie_action_type_counts=sorted_count_items(eyrie_action_type_counts),
            eyrie_turmoil_count=eyrie_turmoil_count,
            eyrie_turmoil_vp_lost=eyrie_turmoil_vp_lost,
            eyrie_turmoil_by_decree_column=sorted_count_items(eyrie_turmoil_by_decree_column),
            eyrie_score_roosts_count=eyrie_score_roosts_count,
            eyrie_score_roosts_vp=eyrie_score_roosts_vp,
            eyrie_roost_builds=eyrie_roost_builds,
            eyrie_roost_losses=eyrie_roost_losses,
            eyrie_roost_losses_on_marquise_turn=eyrie_roost_losses_on_marquise_turn,
            eyrie_cards_added_by_decree_column=tuple(int(value) for value in eyrie_cards_added_by_decree_column),
            mean_eyrie_roosts=eyrie_roost_sum / max(eyrie_roost_samples, 1),
            mean_terminal_eyrie_roosts=float(np.mean(terminal_eyrie_roosts)) if terminal_eyrie_roosts else 0.0,
            mean_terminal_eyrie_vp=float(np.mean(terminal_eyrie_vp)) if terminal_eyrie_vp else 0.0,
            mean_terminal_eyrie_decree_counts=mean_decree_counts(terminal_eyrie_decree_counts),
        ),
    )


@dataclass(frozen=True)
class EyrieTransitionDiagnostics:
    action_type_counts: Counter[int]
    turmoil_count: int
    turmoil_vp_lost: int
    turmoil_by_decree_column: Counter[int]
    score_roosts_count: int
    score_roosts_vp: int
    roost_builds: int
    roost_losses: int
    roost_losses_on_marquise_turn: int
    cards_added_by_decree_column: np.ndarray


def eyrie_transition_diagnostics(
    previous: VecEnvArrays,
    current: VecEnvArrays,
    actions: np.ndarray,
) -> EyrieTransitionDiagnostics:
    action_type_counts: Counter[int] = Counter()
    turmoil_by_decree_column: Counter[int] = Counter()
    cards_added_by_decree_column = np.zeros((4,), dtype=np.int64)
    turmoil_count = 0
    turmoil_vp_lost = 0
    score_roosts_count = 0
    score_roosts_vp = 0
    roost_builds = 0
    roost_losses = 0
    roost_losses_on_marquise_turn = 0

    for row, action_index in enumerate(actions):
        index = int(action_index)
        if index < 0 or index >= len(previous.decisions[row].legal_action_types):
            continue
        action_type = int(previous.decisions[row].legal_action_types[index])
        acting_faction = int(previous.active_factions[row])
        before_vp = int(previous.victory_points[row, FACTION_EYRIE])
        after_vp = int(current.victory_points[row, FACTION_EYRIE])
        roost_delta = int(current.eyrie_roosts[row] - previous.eyrie_roosts[row])

        if acting_faction == FACTION_EYRIE:
            action_type_counts[action_type] += 1

            if action_type == ACTION_TURMOIL:
                turmoil_count += 1
                turmoil_vp_lost += max(0, before_vp - after_vp)
                turmoil_by_decree_column[int(previous.eyrie_current_decree_columns[row])] += 1

            if action_type == ACTION_SCORE_ROOSTS:
                score_roosts_count += 1
                score_roosts_vp += max(0, after_vp - before_vp)

            decree_delta = current.eyrie_decree_counts[row] - previous.eyrie_decree_counts[row]
            cards_added_by_decree_column += np.maximum(decree_delta, 0)

        if roost_delta > 0:
            roost_builds += roost_delta
        elif roost_delta < 0:
            lost = abs(roost_delta)
            roost_losses += lost
            if acting_faction == FACTION_MARQUISE:
                roost_losses_on_marquise_turn += lost

    return EyrieTransitionDiagnostics(
        action_type_counts=action_type_counts,
        turmoil_count=turmoil_count,
        turmoil_vp_lost=turmoil_vp_lost,
        turmoil_by_decree_column=turmoil_by_decree_column,
        score_roosts_count=score_roosts_count,
        score_roosts_vp=score_roosts_vp,
        roost_builds=roost_builds,
        roost_losses=roost_losses,
        roost_losses_on_marquise_turn=roost_losses_on_marquise_turn,
        cards_added_by_decree_column=cards_added_by_decree_column,
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


def mean_decree_counts(values: list[np.ndarray]) -> tuple[float, float, float, float]:
    if len(values) == 0:
        return (0.0, 0.0, 0.0, 0.0)
    stacked = np.stack(values, axis=0)
    return tuple(float(value) for value in stacked.mean(axis=0))  # type: ignore[return-value]


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
        "reward={reward:.4f} decision_steps={length:.2f} tps={tps:.1f} "
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
        "mean_terminal_round_number": round(diagnostics.mean_terminal_round_number, 2),
        "eyrie_action_type_counts": dict(diagnostics.eyrie_action_type_counts),
        "eyrie_turmoil_count": diagnostics.eyrie_turmoil_count,
        "eyrie_turmoil_vp_lost": diagnostics.eyrie_turmoil_vp_lost,
        "eyrie_turmoil_by_decree_column": dict(diagnostics.eyrie_turmoil_by_decree_column),
        "eyrie_score_roosts_count": diagnostics.eyrie_score_roosts_count,
        "eyrie_score_roosts_vp": diagnostics.eyrie_score_roosts_vp,
        "eyrie_roost_builds": diagnostics.eyrie_roost_builds,
        "eyrie_roost_losses": diagnostics.eyrie_roost_losses,
        "eyrie_roost_losses_on_marquise_turn": diagnostics.eyrie_roost_losses_on_marquise_turn,
        "eyrie_cards_added_by_decree_column": diagnostics.eyrie_cards_added_by_decree_column,
        "mean_eyrie_roosts": round(diagnostics.mean_eyrie_roosts, 2),
        "mean_terminal_eyrie_roosts": round(diagnostics.mean_terminal_eyrie_roosts, 2),
        "mean_terminal_eyrie_vp": round(diagnostics.mean_terminal_eyrie_vp, 2),
        "mean_terminal_eyrie_decree_counts": tuple(
            round(value, 2) for value in diagnostics.mean_terminal_eyrie_decree_counts
        ),
    }


if __name__ == "__main__":
    main()
