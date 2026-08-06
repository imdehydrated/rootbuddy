from __future__ import annotations

import pytest
import torch
from torch import nn

from rootbuddy_rl.model import CandidatePolicyValueNet, mask_logits, tensors_from_batch
from rootbuddy_rl.vec_env import batch_to_arrays
from rootbuddy_rl.engine_client import EnvBatch
from test_vec_env import decision


def test_forward_returns_masked_logits_and_values() -> None:
    model = CandidatePolicyValueNet(
        observation_dim=5,
        action_dim=3,
        torso_hidden_dims=(16,),
        head_hidden_dim=8,
    )
    observations = torch.randn(4, 5)
    candidate_actions = torch.randn(4, 3, 3)
    action_mask = torch.tensor(
        [
            [True, True, False],
            [True, False, False],
            [True, True, True],
            [False, True, False],
        ]
    )
    active_factions = torch.tensor([0, 1, 2, 3])

    output = model(observations, candidate_actions, action_mask, active_factions)

    assert output.logits.shape == (4, 3)
    assert output.values.shape == (4,)
    assert torch.isfinite(output.values).all()
    assert output.logits[0, 2] == torch.finfo(output.logits.dtype).min
    assert output.logits[1, 1] == torch.finfo(output.logits.dtype).min
    assert output.logits[3, 0] == torch.finfo(output.logits.dtype).min


def test_act_never_samples_invalid_candidates() -> None:
    torch.manual_seed(707)
    model = CandidatePolicyValueNet(
        observation_dim=4,
        action_dim=2,
        torso_hidden_dims=(8,),
        head_hidden_dim=8,
    )
    observations = torch.randn(3, 4)
    candidate_actions = torch.randn(3, 4, 2)
    action_mask = torch.tensor(
        [
            [True, False, False, False],
            [False, True, False, False],
            [True, True, True, False],
        ]
    )
    active_factions = torch.tensor([0, 1, 2])

    for _ in range(25):
        selection = model.act(observations, candidate_actions, action_mask, active_factions)
        assert selection.actions.shape == (3,)
        assert action_mask[torch.arange(3), selection.actions].all()
        assert selection.log_probs.shape == (3,)
        assert selection.entropy.shape == (3,)
        assert selection.values.shape == (3,)


def test_deterministic_act_chooses_best_valid_logit() -> None:
    model = CandidatePolicyValueNet(
        observation_dim=2,
        action_dim=2,
        torso_hidden_dims=(4,),
        head_hidden_dim=4,
    )
    observations = torch.randn(1, 2)
    candidate_actions = torch.randn(1, 3, 2)
    action_mask = torch.tensor([[True, False, True]])
    active_factions = torch.tensor([0])

    with torch.no_grad():
        output = model(observations, candidate_actions, action_mask, active_factions)
    expected = output.logits.argmax(dim=-1)
    selection = model.act(observations, candidate_actions, action_mask, active_factions, deterministic=True)

    assert torch.equal(selection.actions, expected)


def test_evaluate_actions_matches_act_log_probs() -> None:
    torch.manual_seed(808)
    model = CandidatePolicyValueNet(
        observation_dim=3,
        action_dim=2,
        torso_hidden_dims=(8,),
        head_hidden_dim=8,
    )
    observations = torch.randn(2, 3)
    candidate_actions = torch.randn(2, 2, 2)
    action_mask = torch.ones(2, 2, dtype=torch.bool)
    active_factions = torch.tensor([0, 1])

    selection = model.act(observations, candidate_actions, action_mask, active_factions)
    evaluation = model.evaluate_actions(
        observations,
        candidate_actions,
        action_mask,
        active_factions,
        selection.actions,
    )

    torch.testing.assert_close(evaluation.log_probs, selection.log_probs)
    torch.testing.assert_close(evaluation.values, selection.values)
    assert evaluation.entropy.shape == (2,)


def test_value_head_routes_by_active_faction() -> None:
    model = CandidatePolicyValueNet(
        observation_dim=2,
        action_dim=1,
        torso_hidden_dims=(4,),
        head_hidden_dim=4,
    )
    for faction, head in enumerate(model.value_heads):
        for parameter in head.parameters():
            parameter.data.zero_()
        last_linear = next(module for module in reversed(head) if isinstance(module, nn.Linear))
        last_linear.bias.data.fill_(float(faction))

    observations = torch.ones(4, 2)
    candidate_actions = torch.ones(4, 1, 1)
    action_mask = torch.ones(4, 1, dtype=torch.bool)
    active_factions = torch.tensor([0, 1, 2, 3])

    output = model(observations, candidate_actions, action_mask, active_factions)

    torch.testing.assert_close(output.values, torch.tensor([0.0, 1.0, 2.0, 3.0]))


def test_act_rejects_rows_without_valid_actions() -> None:
    model = CandidatePolicyValueNet(
        observation_dim=2,
        action_dim=1,
        torso_hidden_dims=(4,),
        head_hidden_dim=4,
    )

    with pytest.raises(ValueError, match="at least one valid action"):
        model.act(
            torch.ones(1, 2),
            torch.ones(1, 2, 1),
            torch.zeros(1, 2, dtype=torch.bool),
            torch.zeros(1, dtype=torch.long),
        )


def test_tensors_from_batch_uses_expected_dtypes() -> None:
    arrays = batch_to_arrays(
        EnvBatch(
            decisions=(
                decision(0, observation=[1.0, 2.0], candidates=[[1.0, 0.0]]),
            ),
            observation_length=2,
            action_length=2,
        )
    )

    tensors = tensors_from_batch(arrays)

    assert tensors["observations"].dtype == torch.float32
    assert tensors["candidate_actions"].dtype == torch.float32
    assert tensors["action_mask"].dtype == torch.bool
    assert tensors["active_factions"].dtype == torch.long


def test_mask_logits_replaces_invalid_entries() -> None:
    logits = torch.tensor([[1.0, 2.0]])
    masked = mask_logits(logits, torch.tensor([[True, False]]))

    assert masked[0, 0] == 1.0
    assert masked[0, 1] == torch.finfo(masked.dtype).min
