from __future__ import annotations

import torch

from rootbuddy_rl.model import CandidatePolicyValueNet
from rootbuddy_rl.ppo import (
    PPOBatch,
    PPOConfig,
    compute_gae,
    iter_minibatches,
    normalize_advantages,
    ppo_loss,
    ppo_update,
)


def test_compute_gae_matches_discounted_returns_when_values_zero() -> None:
    rewards = torch.tensor([[1.0], [2.0], [3.0]])
    values = torch.zeros_like(rewards)
    dones = torch.zeros_like(rewards)
    last_values = torch.zeros(1)

    advantages, returns = compute_gae(rewards, values, dones, last_values, gamma=0.5, gae_lambda=1.0)

    expected = torch.tensor([[2.75], [3.5], [3.0]])
    torch.testing.assert_close(advantages, expected)
    torch.testing.assert_close(returns, expected)


def test_compute_gae_resets_at_done_boundaries() -> None:
    rewards = torch.tensor([[1.0], [2.0], [3.0]])
    values = torch.zeros_like(rewards)
    dones = torch.tensor([[0.0], [1.0], [0.0]])
    last_values = torch.tensor([10.0])

    advantages, returns = compute_gae(rewards, values, dones, last_values, gamma=0.5, gae_lambda=1.0)

    expected = torch.tensor([[2.0], [2.0], [8.0]])
    torch.testing.assert_close(advantages, expected)
    torch.testing.assert_close(returns, expected)


def test_normalize_advantages_zero_centers_unit_scales() -> None:
    normalized = normalize_advantages(torch.tensor([1.0, 2.0, 3.0]))

    torch.testing.assert_close(normalized.mean(), torch.tensor(0.0), atol=1e-6, rtol=0.0)
    torch.testing.assert_close(normalized.std(unbiased=False), torch.tensor(1.0), atol=1e-6, rtol=0.0)


def test_ppo_loss_is_finite_and_backpropagates() -> None:
    torch.manual_seed(707)
    model = CandidatePolicyValueNet(
        observation_dim=4,
        action_dim=2,
        torso_hidden_dims=(8,),
        head_hidden_dim=8,
    )
    batch = sample_batch(model)

    loss = ppo_loss(model, batch, PPOConfig(entropy_coef=0.0, normalize_advantages=False))
    loss.total.backward()

    assert torch.isfinite(loss.total)
    assert torch.isfinite(loss.policy)
    assert torch.isfinite(loss.value)
    assert torch.isfinite(loss.entropy)
    assert any(parameter.grad is not None for parameter in model.parameters())


def test_ppo_update_changes_model_parameters() -> None:
    torch.manual_seed(808)
    model = CandidatePolicyValueNet(
        observation_dim=4,
        action_dim=2,
        torso_hidden_dims=(8,),
        head_hidden_dim=8,
    )
    optimizer = torch.optim.Adam(model.parameters(), lr=1e-2)
    batch = sample_batch(model, batch_size=12)
    before = [parameter.detach().clone() for parameter in model.parameters()]

    stats = ppo_update(
        model,
        optimizer,
        batch,
        PPOConfig(epochs=2, minibatch_size=4, entropy_coef=0.0),
    )

    assert stats.updates == 6
    assert torch.isfinite(torch.tensor(stats.loss))
    assert any(not torch.equal(old, new) for old, new in zip(before, model.parameters()))


def test_iter_minibatches_covers_all_rows() -> None:
    model = CandidatePolicyValueNet(
        observation_dim=4,
        action_dim=2,
        torso_hidden_dims=(8,),
        head_hidden_dim=8,
    )
    batch = sample_batch(model, batch_size=10)

    sizes = [minibatch.size for minibatch in iter_minibatches(batch, minibatch_size=4)]

    assert sorted(sizes) == [2, 4, 4]


def sample_batch(model: CandidatePolicyValueNet, batch_size: int = 6) -> PPOBatch:
    observations = torch.randn(batch_size, model.observation_dim)
    candidate_actions = torch.randn(batch_size, 3, model.action_dim)
    action_mask = torch.tensor([[True, True, False]] * batch_size)
    active_factions = torch.arange(batch_size) % model.num_factions
    with torch.no_grad():
        selection = model.act(observations, candidate_actions, action_mask, active_factions)

    rewards = torch.linspace(0.0, 1.0, batch_size)
    returns = selection.values.detach() + rewards
    advantages = returns - selection.values.detach()
    return PPOBatch(
        observations=observations,
        candidate_actions=candidate_actions,
        action_mask=action_mask,
        active_factions=active_factions,
        actions=selection.actions.detach(),
        old_log_probs=selection.log_probs.detach(),
        returns=returns.detach(),
        advantages=advantages.detach(),
    )
