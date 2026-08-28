"""Rollout collection from the Python vector env into PPO batches."""

from __future__ import annotations

from dataclasses import dataclass
from typing import Mapping, Protocol

import numpy as np
import torch

from .model import CandidatePolicyValueNet, tensors_from_batch
from .ppo import PPOBatch
from .train_constants import FACTION_COUNT
from .vec_env import RootBuddyVecEnv, VecEnvArrays


class ActionProvider(Protocol):
    def select_actions(self, batch: VecEnvArrays) -> np.ndarray:
        """Return one action index per env row."""


@dataclass(frozen=True)
class RolloutConfig:
    steps: int
    gamma: float = 0.99
    gae_lambda: float = 0.95
    device: torch.device | str | None = None
    deterministic: bool = False
    stop_on_done: bool = True


@dataclass(frozen=True)
class RolloutStats:
    steps: int
    transitions: int
    completed_episodes: int
    mean_reward: float
    mean_game_length: float
    per_faction_win_rate: tuple[float, ...]
    per_faction_terminal_vp: tuple[float, ...]


@dataclass(frozen=True)
class RolloutResult:
    batch: PPOBatch
    stats: RolloutStats
    last_env_batch: VecEnvArrays


def collect_rollout(
    env: RootBuddyVecEnv,
    model: CandidatePolicyValueNet,
    config: RolloutConfig,
    *,
    initial_batch: VecEnvArrays | None = None,
    opponent_agents: Mapping[int, ActionProvider] | None = None,
) -> RolloutResult:
    if config.steps <= 0:
        raise ValueError("rollout steps must be positive")

    model_was_training = model.training
    model.eval()
    device = config.device
    current = initial_batch if initial_batch is not None else env.reset()
    if bool(current.dones.any()) and not config.stop_on_done:
        current = env.reset(env_indices=env_indices_for_rows(current, np.flatnonzero(current.dones)))

    observations: list[torch.Tensor] = []
    candidate_actions: list[torch.Tensor] = []
    action_masks: list[torch.Tensor] = []
    active_factions: list[torch.Tensor] = []
    actions: list[torch.Tensor] = []
    old_log_probs: list[torch.Tensor] = []
    values: list[torch.Tensor] = []
    rewards: list[torch.Tensor] = []
    dones: list[torch.Tensor] = []
    learning_masks: list[torch.Tensor] = []
    terminal_steps: list[int] = []
    terminal_vp: list[np.ndarray] = []
    winner_counts = [0 for _ in range(FACTION_COUNT)]

    for _ in range(config.steps):
        if (current.candidate_counts <= 0).any():
            raise ValueError("cannot collect from env batch with no legal actions")

        tensors = tensors_from_batch(current, device=device)
        with torch.no_grad():
            selection = model.act(
                tensors["observations"],
                tensors["candidate_actions"],
                tensors["action_mask"],
                tensors["active_factions"],
                deterministic=config.deterministic,
            )

        action_indices, learning_mask = choose_rollout_actions(
            current,
            selection.actions.detach().cpu().numpy().astype(np.int64),
            opponent_agents=opponent_agents,
        )
        next_batch = env.step(action_indices.tolist())

        observations.append(tensors["observations"].detach().cpu())
        candidate_actions.append(tensors["candidate_actions"].detach().cpu())
        action_masks.append(tensors["action_mask"].detach().cpu())
        active_factions.append(tensors["active_factions"].detach().cpu())
        actions.append(torch.as_tensor(action_indices, dtype=torch.long))
        old_log_probs.append(selection.log_probs.detach().cpu())
        values.append(selection.values.detach().cpu())
        rewards.append(torch.as_tensor(next_batch.rewards, dtype=torch.float32))
        dones.append(torch.as_tensor(next_batch.dones, dtype=torch.float32))
        learning_masks.append(torch.as_tensor(learning_mask, dtype=torch.bool))

        done_rows = np.flatnonzero(next_batch.dones)
        if done_rows.size > 0:
            record_terminal_outcomes(
                next_batch,
                done_rows,
                terminal_steps=terminal_steps,
                terminal_vp=terminal_vp,
                winner_counts=winner_counts,
            )

        current = next_batch
        if config.stop_on_done and bool(next_batch.dones.any()):
            break
        if done_rows.size > 0:
            current = env.reset(env_indices=env_indices_for_rows(next_batch, done_rows))

    if model_was_training:
        model.train()

    if len(observations) == 0:
        raise ValueError("rollout produced no transitions")

    rewards_by_time = torch.stack(rewards)
    values_by_time = torch.stack(values)
    dones_by_time = torch.stack(dones)
    active_factions_by_time = torch.stack(active_factions)
    last_values = bootstrap_values(model, current, device=device).cpu()
    last_active_factions = torch.as_tensor(current.active_factions, dtype=torch.long)
    advantages, returns = compute_turn_based_gae(
        rewards_by_time,
        values_by_time,
        dones_by_time,
        active_factions_by_time,
        last_values,
        last_active_factions,
        gamma=config.gamma,
        gae_lambda=config.gae_lambda,
    )

    learning_mask = torch.cat(learning_masks, dim=0)
    if not bool(learning_mask.any()):
        raise ValueError("rollout produced no live-model transitions")

    batch = PPOBatch(
        observations=torch.cat(observations, dim=0)[learning_mask],
        candidate_actions=flatten_padded_candidate_actions(candidate_actions)[learning_mask],
        action_mask=flatten_padded_action_masks(action_masks)[learning_mask],
        active_factions=torch.cat(active_factions, dim=0)[learning_mask],
        actions=torch.cat(actions, dim=0)[learning_mask],
        old_log_probs=torch.cat(old_log_probs, dim=0)[learning_mask],
        returns=returns.reshape(-1)[learning_mask],
        advantages=advantages.reshape(-1)[learning_mask],
    )
    completed_episodes = len(terminal_steps)
    stats = RolloutStats(
        steps=len(rewards),
        transitions=batch.size,
        completed_episodes=completed_episodes,
        mean_reward=float(rewards_by_time.mean().item()),
        mean_game_length=float(np.mean(terminal_steps)) if terminal_steps else 0.0,
        per_faction_win_rate=tuple(count / completed_episodes for count in winner_counts)
        if completed_episodes > 0
        else tuple(0.0 for _ in range(FACTION_COUNT)),
        per_faction_terminal_vp=mean_terminal_vp(terminal_vp),
    )
    return RolloutResult(batch=batch, stats=stats, last_env_batch=current)


def env_indices_for_rows(batch: VecEnvArrays, rows: np.ndarray) -> list[int]:
    return [int(batch.decisions[int(row)].env_index) for row in rows]


def record_terminal_outcomes(
    batch: VecEnvArrays,
    rows: np.ndarray,
    *,
    terminal_steps: list[int],
    terminal_vp: list[np.ndarray],
    winner_counts: list[int],
) -> None:
    for row in rows:
        row_index = int(row)
        terminal_steps.append(int(batch.steps[row_index]))
        terminal_vp.append(batch.victory_points[row_index].astype(np.float32))
        if bool(batch.truncations[row_index]):
            continue
        winner = int(batch.winners[row_index])
        if 0 <= winner < len(winner_counts):
            winner_counts[winner] += 1


def mean_terminal_vp(values: list[np.ndarray]) -> tuple[float, ...]:
    if len(values) == 0:
        return tuple(0.0 for _ in range(FACTION_COUNT))
    stacked = np.stack(values, axis=0)
    return tuple(float(value) for value in stacked.mean(axis=0))


def compute_turn_based_gae(
    rewards: torch.Tensor,
    values: torch.Tensor,
    dones: torch.Tensor,
    active_factions: torch.Tensor,
    last_values: torch.Tensor,
    last_active_factions: torch.Tensor,
    *,
    gamma: float = 0.99,
    gae_lambda: float = 0.95,
) -> tuple[torch.Tensor, torch.Tensor]:
    if rewards.shape != values.shape or rewards.shape != dones.shape or rewards.shape != active_factions.shape:
        raise ValueError("rewards, values, dones, and active_factions must have matching [time, env] shapes")
    if rewards.ndim != 2:
        raise ValueError("turn-based GAE inputs must have shape [time, env]")
    if last_values.shape != rewards.shape[1:] or last_active_factions.shape != rewards.shape[1:]:
        raise ValueError("last_values and last_active_factions must have shape [env]")

    advantages = torch.zeros_like(rewards)
    next_advantages = rewards.new_zeros((rewards.shape[1], FACTION_COUNT))
    next_values = rewards.new_zeros((rewards.shape[1], FACTION_COUNT))
    for env_index, faction in enumerate(last_active_factions.to(dtype=torch.long).tolist()):
        if 0 <= faction < FACTION_COUNT:
            next_values[env_index, faction] = last_values[env_index]

    for step in reversed(range(rewards.shape[0])):
        for env_index in range(rewards.shape[1]):
            faction = int(active_factions[step, env_index].item())
            if faction < 0 or faction >= FACTION_COUNT:
                raise ValueError("active_factions contains invalid faction id")
            if bool(dones[step, env_index].item()):
                next_advantages[env_index, :] = 0
                next_values[env_index, :] = 0
            delta = rewards[step, env_index] + gamma * next_values[env_index, faction] - values[step, env_index]
            advantage = delta + gamma * gae_lambda * next_advantages[env_index, faction]
            advantages[step, env_index] = advantage
            next_advantages[env_index, faction] = advantage
            next_values[env_index, faction] = values[step, env_index]

    returns = advantages + values
    return advantages, returns


def choose_rollout_actions(
    batch: VecEnvArrays,
    live_actions: np.ndarray,
    *,
    opponent_agents: Mapping[int, ActionProvider] | None = None,
) -> tuple[np.ndarray, np.ndarray]:
    if live_actions.shape != (batch.num_envs,):
        raise ValueError("live_actions must have one action per env")
    actions = live_actions.copy()
    learning_mask = np.ones((batch.num_envs,), dtype=np.bool_)
    if not opponent_agents:
        return actions, learning_mask

    proposals: dict[int, np.ndarray] = {}
    for faction, agent in opponent_agents.items():
        rows = batch.active_factions == int(faction)
        if not bool(rows.any()):
            continue
        proposal_key = id(agent)
        if proposal_key not in proposals:
            proposal = agent.select_actions(batch)
            if proposal.shape != (batch.num_envs,):
                raise ValueError("opponent agent returned wrong action shape")
            proposals[proposal_key] = proposal.astype(np.int64, copy=False)
        actions[rows] = proposals[proposal_key][rows]
        learning_mask[rows] = False
    if not bool(learning_mask.any()):
        learning_mask[0] = True
        actions[0] = live_actions[0]
    if np.any(actions < 0):
        raise ValueError("rollout action selection left invalid rows")
    return actions, learning_mask


def bootstrap_values(
    model: CandidatePolicyValueNet,
    env_batch: VecEnvArrays,
    *,
    device: torch.device | str | None = None,
) -> torch.Tensor:
    if bool(env_batch.dones.all()):
        return torch.zeros((env_batch.num_envs,), dtype=torch.float32, device=device)
    tensors = tensors_from_batch(env_batch, device=device)
    with torch.no_grad():
        values = model(
            tensors["observations"],
            tensors["candidate_actions"],
            tensors["action_mask"],
            tensors["active_factions"],
        ).values
    done_mask = torch.as_tensor(env_batch.dones, dtype=torch.bool, device=values.device)
    return values.masked_fill(done_mask, 0.0)


def flatten_padded_candidate_actions(items: list[torch.Tensor]) -> torch.Tensor:
    if len(items) == 0:
        raise ValueError("cannot flatten empty candidate action list")
    max_candidates = max(tensor.shape[1] for tensor in items)
    action_dim = items[0].shape[2]
    padded = [pad_candidates(tensor, max_candidates=max_candidates, action_dim=action_dim) for tensor in items]
    return torch.cat(padded, dim=0)


def flatten_padded_action_masks(items: list[torch.Tensor]) -> torch.Tensor:
    if len(items) == 0:
        raise ValueError("cannot flatten empty action mask list")
    max_candidates = max(tensor.shape[1] for tensor in items)
    padded = [pad_action_mask(tensor, max_candidates=max_candidates) for tensor in items]
    return torch.cat(padded, dim=0)


def pad_candidates(tensor: torch.Tensor, *, max_candidates: int, action_dim: int) -> torch.Tensor:
    if tensor.shape[1] == max_candidates:
        return tensor
    padded = tensor.new_zeros((tensor.shape[0], max_candidates, action_dim))
    padded[:, : tensor.shape[1], :] = tensor
    return padded


def pad_action_mask(tensor: torch.Tensor, *, max_candidates: int) -> torch.Tensor:
    if tensor.shape[1] == max_candidates:
        return tensor
    padded = torch.zeros((tensor.shape[0], max_candidates), dtype=torch.bool, device=tensor.device)
    padded[:, : tensor.shape[1]] = tensor
    return padded
