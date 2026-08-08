"""PPO loss and update utilities for candidate-action policies."""

from __future__ import annotations

from dataclasses import dataclass
from typing import Iterator

import torch
from torch import nn

from .model import CandidatePolicyValueNet


@dataclass(frozen=True)
class PPOConfig:
    clip_range: float = 0.2
    value_coef: float = 0.5
    entropy_coef: float = 0.01
    max_grad_norm: float = 0.5
    epochs: int = 4
    minibatch_size: int = 256
    normalize_advantages: bool = True


@dataclass(frozen=True)
class PPOBatch:
    observations: torch.Tensor
    candidate_actions: torch.Tensor
    action_mask: torch.Tensor
    active_factions: torch.Tensor
    actions: torch.Tensor
    old_log_probs: torch.Tensor
    returns: torch.Tensor
    advantages: torch.Tensor

    def __post_init__(self) -> None:
        batch_size = self.observations.shape[0]
        checks = {
            "candidate_actions": self.candidate_actions.shape[0],
            "action_mask": self.action_mask.shape[0],
            "active_factions": self.active_factions.shape[0],
            "actions": self.actions.shape[0],
            "old_log_probs": self.old_log_probs.shape[0],
            "returns": self.returns.shape[0],
            "advantages": self.advantages.shape[0],
        }
        for name, size in checks.items():
            if size != batch_size:
                raise ValueError(f"{name} batch size {size} does not match observations {batch_size}")
        if self.candidate_actions.ndim != 3:
            raise ValueError("candidate_actions must have shape [batch, max_candidates, action_dim]")
        if self.action_mask.shape != self.candidate_actions.shape[:2]:
            raise ValueError("action_mask shape must match candidate action rows")

    @property
    def size(self) -> int:
        return int(self.observations.shape[0])

    def to(self, device: torch.device | str) -> PPOBatch:
        return PPOBatch(
            observations=self.observations.to(device),
            candidate_actions=self.candidate_actions.to(device),
            action_mask=self.action_mask.to(device),
            active_factions=self.active_factions.to(device),
            actions=self.actions.to(device),
            old_log_probs=self.old_log_probs.to(device),
            returns=self.returns.to(device),
            advantages=self.advantages.to(device),
        )

    def index_select(self, indices: torch.Tensor) -> PPOBatch:
        return PPOBatch(
            observations=self.observations[indices],
            candidate_actions=self.candidate_actions[indices],
            action_mask=self.action_mask[indices],
            active_factions=self.active_factions[indices],
            actions=self.actions[indices],
            old_log_probs=self.old_log_probs[indices],
            returns=self.returns[indices],
            advantages=self.advantages[indices],
        )


@dataclass(frozen=True)
class PPOLoss:
    total: torch.Tensor
    policy: torch.Tensor
    value: torch.Tensor
    entropy: torch.Tensor
    approx_kl: torch.Tensor
    clip_fraction: torch.Tensor


@dataclass(frozen=True)
class PPOUpdateStats:
    loss: float
    policy_loss: float
    value_loss: float
    entropy: float
    approx_kl: float
    clip_fraction: float
    updates: int


def compute_gae(
    rewards: torch.Tensor,
    values: torch.Tensor,
    dones: torch.Tensor,
    last_values: torch.Tensor,
    *,
    gamma: float = 0.99,
    gae_lambda: float = 0.95,
) -> tuple[torch.Tensor, torch.Tensor]:
    if rewards.shape != values.shape or rewards.shape != dones.shape:
        raise ValueError("rewards, values, and dones must have matching [time, env] shapes")
    if rewards.ndim != 2:
        raise ValueError("rewards, values, and dones must have shape [time, env]")
    if last_values.shape != rewards.shape[1:]:
        raise ValueError("last_values must have shape [env]")

    advantages = torch.zeros_like(rewards)
    next_advantage = torch.zeros_like(last_values)
    next_values = last_values
    for step in reversed(range(rewards.shape[0])):
        not_done = 1.0 - dones[step].to(dtype=rewards.dtype)
        delta = rewards[step] + gamma * next_values * not_done - values[step]
        next_advantage = delta + gamma * gae_lambda * not_done * next_advantage
        advantages[step] = next_advantage
        next_values = values[step]
    returns = advantages + values
    return advantages, returns


def normalize_advantages(advantages: torch.Tensor, eps: float = 1e-8) -> torch.Tensor:
    if advantages.numel() <= 1:
        return advantages
    return (advantages - advantages.mean()) / (advantages.std(unbiased=False) + eps)


def ppo_loss(
    model: CandidatePolicyValueNet,
    batch: PPOBatch,
    config: PPOConfig = PPOConfig(),
) -> PPOLoss:
    advantages = batch.advantages
    if config.normalize_advantages:
        advantages = normalize_advantages(advantages)

    evaluation = model.evaluate_actions(
        batch.observations,
        batch.candidate_actions,
        batch.action_mask,
        batch.active_factions,
        batch.actions,
    )
    log_ratio = evaluation.log_probs - batch.old_log_probs
    ratio = torch.exp(log_ratio)
    clipped_ratio = torch.clamp(ratio, 1.0 - config.clip_range, 1.0 + config.clip_range)
    policy_loss = -torch.min(ratio * advantages, clipped_ratio * advantages).mean()
    value_loss = 0.5 * torch.square(evaluation.values - batch.returns).mean()
    entropy = evaluation.entropy.mean()
    total = policy_loss + config.value_coef * value_loss - config.entropy_coef * entropy

    with torch.no_grad():
        approx_kl = (batch.old_log_probs - evaluation.log_probs).mean()
        clip_fraction = ((ratio - 1.0).abs() > config.clip_range).to(dtype=torch.float32).mean()

    return PPOLoss(
        total=total,
        policy=policy_loss,
        value=value_loss,
        entropy=entropy,
        approx_kl=approx_kl,
        clip_fraction=clip_fraction,
    )


def ppo_update(
    model: CandidatePolicyValueNet,
    optimizer: torch.optim.Optimizer,
    batch: PPOBatch,
    config: PPOConfig = PPOConfig(),
) -> PPOUpdateStats:
    if batch.size == 0:
        raise ValueError("cannot update PPO with an empty batch")

    totals: list[PPOLoss] = []
    for _ in range(config.epochs):
        for minibatch in iter_minibatches(batch, config.minibatch_size):
            optimizer.zero_grad(set_to_none=True)
            loss = ppo_loss(model, minibatch, config)
            loss.total.backward()
            if config.max_grad_norm > 0:
                nn.utils.clip_grad_norm_(model.parameters(), config.max_grad_norm)
            optimizer.step()
            totals.append(detach_loss(loss))

    return summarize_losses(totals)


def iter_minibatches(batch: PPOBatch, minibatch_size: int) -> Iterator[PPOBatch]:
    if minibatch_size <= 0:
        raise ValueError("minibatch_size must be positive")
    permutation = torch.randperm(batch.size, device=batch.observations.device)
    for start in range(0, batch.size, minibatch_size):
        yield batch.index_select(permutation[start : start + minibatch_size])


def detach_loss(loss: PPOLoss) -> PPOLoss:
    return PPOLoss(
        total=loss.total.detach(),
        policy=loss.policy.detach(),
        value=loss.value.detach(),
        entropy=loss.entropy.detach(),
        approx_kl=loss.approx_kl.detach(),
        clip_fraction=loss.clip_fraction.detach(),
    )


def summarize_losses(losses: list[PPOLoss]) -> PPOUpdateStats:
    if len(losses) == 0:
        raise ValueError("cannot summarize empty PPO losses")

    def mean_of(field: str) -> float:
        return float(torch.stack([getattr(loss, field) for loss in losses]).mean().item())

    return PPOUpdateStats(
        loss=mean_of("total"),
        policy_loss=mean_of("policy"),
        value_loss=mean_of("value"),
        entropy=mean_of("entropy"),
        approx_kl=mean_of("approx_kl"),
        clip_fraction=mean_of("clip_fraction"),
        updates=len(losses),
    )
