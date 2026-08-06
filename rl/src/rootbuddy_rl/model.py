"""Candidate-action policy/value network for RootBuddy PPO."""

from __future__ import annotations

from dataclasses import dataclass
from typing import Sequence

import torch
from torch import nn
from torch.distributions import Categorical

from .vec_env import VecEnvArrays


@dataclass(frozen=True)
class ModelOutput:
    logits: torch.Tensor
    values: torch.Tensor


@dataclass(frozen=True)
class ActionSelection:
    actions: torch.Tensor
    log_probs: torch.Tensor
    entropy: torch.Tensor
    values: torch.Tensor
    logits: torch.Tensor


@dataclass(frozen=True)
class ActionEvaluation:
    log_probs: torch.Tensor
    entropy: torch.Tensor
    values: torch.Tensor
    logits: torch.Tensor


class CandidatePolicyValueNet(nn.Module):
    def __init__(
        self,
        observation_dim: int,
        action_dim: int,
        *,
        num_factions: int = 4,
        torso_hidden_dims: Sequence[int] = (256, 256),
        head_hidden_dim: int = 128,
    ) -> None:
        super().__init__()
        if observation_dim <= 0:
            raise ValueError("observation_dim must be positive")
        if action_dim <= 0:
            raise ValueError("action_dim must be positive")
        if num_factions <= 0:
            raise ValueError("num_factions must be positive")

        self.observation_dim = observation_dim
        self.action_dim = action_dim
        self.num_factions = num_factions
        self.embedding_dim = torso_hidden_dims[-1] if len(torso_hidden_dims) > 0 else observation_dim

        self.torso = mlp(observation_dim, torso_hidden_dims, self.embedding_dim)
        policy_input_dim = self.embedding_dim + action_dim
        self.policy_heads = nn.ModuleList(
            mlp(policy_input_dim, (head_hidden_dim,), 1) for _ in range(num_factions)
        )
        self.value_heads = nn.ModuleList(
            mlp(self.embedding_dim, (head_hidden_dim,), 1) for _ in range(num_factions)
        )

    def forward(
        self,
        observations: torch.Tensor,
        candidate_actions: torch.Tensor,
        action_mask: torch.Tensor,
        active_factions: torch.Tensor,
    ) -> ModelOutput:
        self._validate_inputs(observations, candidate_actions, action_mask, active_factions)

        batch_size, max_candidates, _ = candidate_actions.shape
        embeddings = self.torso(observations)
        raw_logits = observations.new_empty((batch_size, max_candidates))
        values = observations.new_empty((batch_size,))
        active_factions = active_factions.to(dtype=torch.long)

        for faction in range(self.num_factions):
            rows = active_factions == faction
            if not torch.any(rows):
                continue

            faction_embeddings = embeddings[rows]
            faction_candidates = candidate_actions[rows]
            repeated_embeddings = faction_embeddings.unsqueeze(1).expand(-1, max_candidates, -1)
            policy_inputs = torch.cat((repeated_embeddings, faction_candidates), dim=-1)
            raw_logits[rows] = self.policy_heads[faction](policy_inputs).squeeze(-1)
            values[rows] = self.value_heads[faction](faction_embeddings).squeeze(-1)

        masked_logits = mask_logits(raw_logits, action_mask)
        return ModelOutput(logits=masked_logits, values=values)

    @torch.no_grad()
    def act(
        self,
        observations: torch.Tensor,
        candidate_actions: torch.Tensor,
        action_mask: torch.Tensor,
        active_factions: torch.Tensor,
        *,
        deterministic: bool = False,
    ) -> ActionSelection:
        output = self.forward(observations, candidate_actions, action_mask, active_factions)
        distribution = masked_categorical(output.logits, action_mask)
        if deterministic:
            actions = output.logits.argmax(dim=-1)
        else:
            actions = distribution.sample()
        return ActionSelection(
            actions=actions,
            log_probs=distribution.log_prob(actions),
            entropy=distribution.entropy(),
            values=output.values,
            logits=output.logits,
        )

    def evaluate_actions(
        self,
        observations: torch.Tensor,
        candidate_actions: torch.Tensor,
        action_mask: torch.Tensor,
        active_factions: torch.Tensor,
        actions: torch.Tensor,
    ) -> ActionEvaluation:
        output = self.forward(observations, candidate_actions, action_mask, active_factions)
        distribution = masked_categorical(output.logits, action_mask)
        actions = actions.to(dtype=torch.long)
        return ActionEvaluation(
            log_probs=distribution.log_prob(actions),
            entropy=distribution.entropy(),
            values=output.values,
            logits=output.logits,
        )

    def _validate_inputs(
        self,
        observations: torch.Tensor,
        candidate_actions: torch.Tensor,
        action_mask: torch.Tensor,
        active_factions: torch.Tensor,
    ) -> None:
        if observations.ndim != 2:
            raise ValueError("observations must have shape [batch, observation_dim]")
        if candidate_actions.ndim != 3:
            raise ValueError("candidate_actions must have shape [batch, max_candidates, action_dim]")
        if action_mask.ndim != 2:
            raise ValueError("action_mask must have shape [batch, max_candidates]")
        if active_factions.ndim != 1:
            raise ValueError("active_factions must have shape [batch]")
        if observations.shape[1] != self.observation_dim:
            raise ValueError(f"observation dim {observations.shape[1]} does not match {self.observation_dim}")
        if candidate_actions.shape[2] != self.action_dim:
            raise ValueError(f"action dim {candidate_actions.shape[2]} does not match {self.action_dim}")
        if observations.shape[0] != candidate_actions.shape[0]:
            raise ValueError("observation and candidate batch sizes must match")
        if action_mask.shape != candidate_actions.shape[:2]:
            raise ValueError("action_mask shape must match candidate action rows")
        if active_factions.shape[0] != observations.shape[0]:
            raise ValueError("active_factions length must match batch size")
        if torch.any(active_factions < 0) or torch.any(active_factions >= self.num_factions):
            raise ValueError("active_factions contains faction outside model head range")


def mlp(input_dim: int, hidden_dims: Sequence[int], output_dim: int) -> nn.Sequential:
    layers: list[nn.Module] = []
    previous_dim = input_dim
    for hidden_dim in hidden_dims:
        layers.append(nn.Linear(previous_dim, hidden_dim))
        layers.append(nn.Tanh())
        previous_dim = hidden_dim
    layers.append(nn.Linear(previous_dim, output_dim))
    return nn.Sequential(*layers)


def mask_logits(logits: torch.Tensor, action_mask: torch.Tensor) -> torch.Tensor:
    mask = action_mask.to(dtype=torch.bool)
    invalid_value = torch.finfo(logits.dtype).min
    return logits.masked_fill(~mask, invalid_value)


def masked_categorical(logits: torch.Tensor, action_mask: torch.Tensor) -> Categorical:
    mask = action_mask.to(dtype=torch.bool)
    if not torch.all(mask.any(dim=-1)):
        raise ValueError("each batch row must contain at least one valid action")
    return Categorical(logits=mask_logits(logits, mask))


def tensors_from_batch(batch: VecEnvArrays, *, device: torch.device | str | None = None) -> dict[str, torch.Tensor]:
    return {
        "observations": torch.as_tensor(batch.observations, dtype=torch.float32, device=device),
        "candidate_actions": torch.as_tensor(batch.candidate_actions, dtype=torch.float32, device=device),
        "action_mask": torch.as_tensor(batch.action_mask, dtype=torch.bool, device=device),
        "active_factions": torch.as_tensor(batch.active_factions, dtype=torch.long, device=device),
    }
