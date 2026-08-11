"""Rollout collection from the Python vector env into PPO batches."""

from __future__ import annotations

from dataclasses import dataclass

import torch

from .model import CandidatePolicyValueNet, tensors_from_batch
from .ppo import PPOBatch, compute_gae
from .vec_env import RootBuddyVecEnv, VecEnvArrays


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


@dataclass(frozen=True)
class RolloutResult:
    batch: PPOBatch
    stats: RolloutStats
    last_env_batch: VecEnvArrays


def collect_rollout(
    env: RootBuddyVecEnv,
    model: CandidatePolicyValueNet,
    config: RolloutConfig,
) -> RolloutResult:
    if config.steps <= 0:
        raise ValueError("rollout steps must be positive")

    model_was_training = model.training
    model.eval()
    device = config.device
    current = env.reset()

    observations: list[torch.Tensor] = []
    candidate_actions: list[torch.Tensor] = []
    action_masks: list[torch.Tensor] = []
    active_factions: list[torch.Tensor] = []
    actions: list[torch.Tensor] = []
    old_log_probs: list[torch.Tensor] = []
    values: list[torch.Tensor] = []
    rewards: list[torch.Tensor] = []
    dones: list[torch.Tensor] = []

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

        next_batch = env.step(selection.actions.detach().cpu().tolist())

        observations.append(tensors["observations"].detach().cpu())
        candidate_actions.append(tensors["candidate_actions"].detach().cpu())
        action_masks.append(tensors["action_mask"].detach().cpu())
        active_factions.append(tensors["active_factions"].detach().cpu())
        actions.append(selection.actions.detach().cpu())
        old_log_probs.append(selection.log_probs.detach().cpu())
        values.append(selection.values.detach().cpu())
        rewards.append(torch.as_tensor(next_batch.rewards, dtype=torch.float32))
        dones.append(torch.as_tensor(next_batch.dones, dtype=torch.float32))

        current = next_batch
        if config.stop_on_done and bool(next_batch.dones.any()):
            break

    if model_was_training:
        model.train()

    if len(observations) == 0:
        raise ValueError("rollout produced no transitions")

    rewards_by_time = torch.stack(rewards)
    values_by_time = torch.stack(values)
    dones_by_time = torch.stack(dones)
    last_values = bootstrap_values(model, current, device=device).cpu()
    advantages, returns = compute_gae(
        rewards_by_time,
        values_by_time,
        dones_by_time,
        last_values,
        gamma=config.gamma,
        gae_lambda=config.gae_lambda,
    )

    batch = PPOBatch(
        observations=torch.cat(observations, dim=0),
        candidate_actions=flatten_padded_candidate_actions(candidate_actions),
        action_mask=flatten_padded_action_masks(action_masks),
        active_factions=torch.cat(active_factions, dim=0),
        actions=torch.cat(actions, dim=0),
        old_log_probs=torch.cat(old_log_probs, dim=0),
        returns=returns.reshape(-1),
        advantages=advantages.reshape(-1),
    )
    stats = RolloutStats(
        steps=len(rewards),
        transitions=batch.size,
        completed_episodes=int(dones_by_time.sum().item()),
        mean_reward=float(rewards_by_time.mean().item()),
    )
    return RolloutResult(batch=batch, stats=stats, last_env_batch=current)


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
