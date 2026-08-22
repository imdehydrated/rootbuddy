from __future__ import annotations

import numpy as np
import torch

from rootbuddy_rl.engine_client import EnvBatch
from rootbuddy_rl.model import CandidatePolicyValueNet
from rootbuddy_rl.rollout import (
    RolloutConfig,
    bootstrap_values,
    collect_rollout,
    flatten_padded_action_masks,
    flatten_padded_candidate_actions,
)
from rootbuddy_rl.vec_env import VecEnvArrays, batch_to_arrays
from test_vec_env import decision


class ConstantAgent:
    def __init__(self, action: int) -> None:
        self.action = action

    def select_actions(self, batch: VecEnvArrays) -> np.ndarray:
        return np.full((batch.num_envs,), self.action, dtype=np.int64)


class FakeVecEnv:
    def __init__(self, reset_batch: VecEnvArrays, step_batches: list[VecEnvArrays]) -> None:
        self.reset_batch = reset_batch
        self.step_batches = step_batches
        self.step_actions: list[list[int]] = []

    def reset(self) -> VecEnvArrays:
        return self.reset_batch

    def step(self, action_indices: list[int]) -> VecEnvArrays:
        self.step_actions.append(action_indices)
        if len(self.step_batches) == 0:
            raise AssertionError("fake env received more steps than expected")
        return self.step_batches.pop(0)


def test_collect_rollout_builds_ppo_batch_and_stats() -> None:
    torch.manual_seed(707)
    model = CandidatePolicyValueNet(
        observation_dim=3,
        action_dim=2,
        torso_hidden_dims=(8,),
        head_hidden_dim=8,
    )
    env = FakeVecEnv(
        reset_batch=make_batch(
            rewards=[0.0, 0.0],
            dones=[False, False],
            candidate_counts=[2, 1],
        ),
        step_batches=[
            make_batch(
                rewards=[1.0, 0.0],
                dones=[False, False],
                candidate_counts=[3, 2],
            ),
            make_batch(
                rewards=[0.0, 2.0],
                dones=[False, True],
                candidate_counts=[1, 1],
            ),
        ],
    )

    result = collect_rollout(
        env,  # type: ignore[arg-type]
        model,
        RolloutConfig(steps=4, gamma=0.9, gae_lambda=0.95),
    )

    assert result.stats.steps == 2
    assert result.stats.transitions == 4
    assert result.stats.completed_episodes == 1
    assert result.batch.observations.shape == (4, 3)
    assert result.batch.candidate_actions.shape == (4, 3, 2)
    assert result.batch.action_mask.shape == (4, 3)
    assert result.batch.actions.shape == (4,)
    assert result.batch.old_log_probs.shape == (4,)
    assert result.batch.returns.shape == (4,)
    assert result.batch.advantages.shape == (4,)
    assert len(env.step_actions) == 2
    for action_row in env.step_actions:
        assert len(action_row) == 2


def test_collect_rollout_can_continue_without_stop_on_done() -> None:
    model = CandidatePolicyValueNet(
        observation_dim=3,
        action_dim=2,
        torso_hidden_dims=(8,),
        head_hidden_dim=8,
    )
    env = FakeVecEnv(
        reset_batch=make_batch([0.0], [False], [1]),
        step_batches=[
            make_batch([1.0], [True], [1]),
            make_batch([0.0], [False], [1]),
        ],
    )

    result = collect_rollout(
        env,  # type: ignore[arg-type]
        model,
        RolloutConfig(steps=2, stop_on_done=False),
    )

    assert result.stats.steps == 2
    assert result.stats.completed_episodes == 1
    assert result.batch.size == 2


def test_collect_rollout_uses_opponents_and_filters_training_rows() -> None:
    model = CandidatePolicyValueNet(
        observation_dim=3,
        action_dim=2,
        torso_hidden_dims=(8,),
        head_hidden_dim=8,
    )
    env = FakeVecEnv(
        reset_batch=make_batch(
            rewards=[0.0, 0.0],
            dones=[False, False],
            candidate_counts=[2, 2],
            active_factions=[0, 2],
        ),
        step_batches=[
            make_batch(
                rewards=[1.0, 3.0],
                dones=[False, True],
                candidate_counts=[2, 2],
                active_factions=[0, 2],
            ),
        ],
    )

    result = collect_rollout(
        env,  # type: ignore[arg-type]
        model,
        RolloutConfig(steps=1),
        opponent_agents={2: ConstantAgent(0)},
    )

    assert len(env.step_actions) == 1
    assert env.step_actions[0][1] == 0
    assert result.batch.size == 1
    assert result.batch.active_factions.tolist() == [0]


def test_collect_rollout_keeps_one_live_row_when_all_active_rows_are_opponents() -> None:
    model = CandidatePolicyValueNet(
        observation_dim=3,
        action_dim=2,
        torso_hidden_dims=(8,),
        head_hidden_dim=8,
    )
    env = FakeVecEnv(
        reset_batch=make_batch([0.0], [False], [2], active_factions=[0]),
        step_batches=[make_batch([1.0], [False], [2], active_factions=[0])],
    )

    result = collect_rollout(
        env,  # type: ignore[arg-type]
        model,
        RolloutConfig(steps=1),
        opponent_agents={0: ConstantAgent(0)},
    )

    assert result.batch.size == 1
    assert result.batch.active_factions.tolist() == [0]


def test_bootstrap_values_zeroes_terminal_rows() -> None:
    model = CandidatePolicyValueNet(
        observation_dim=3,
        action_dim=2,
        torso_hidden_dims=(8,),
        head_hidden_dim=8,
    )
    batch = make_batch(
        rewards=[0.0, 0.0],
        dones=[False, True],
        candidate_counts=[2, 1],
    )

    values = bootstrap_values(model, batch)

    assert values.shape == (2,)
    assert values[1] == 0


def test_flatten_padded_rollout_tensors_uses_global_candidate_width() -> None:
    first_candidates = torch.ones(2, 1, 3)
    second_candidates = torch.ones(2, 4, 3)
    first_mask = torch.ones(2, 1, dtype=torch.bool)
    second_mask = torch.ones(2, 4, dtype=torch.bool)

    candidates = flatten_padded_candidate_actions([first_candidates, second_candidates])
    masks = flatten_padded_action_masks([first_mask, second_mask])

    assert candidates.shape == (4, 4, 3)
    assert masks.shape == (4, 4)
    assert not masks[0, 1:].any()
    np.testing.assert_array_equal(candidates[0, 1:].numpy(), np.zeros((3, 3), dtype=np.float32))


def make_batch(
    rewards: list[float],
    dones: list[bool],
    candidate_counts: list[int],
    active_factions: list[int] | None = None,
) -> VecEnvArrays:
    decisions = []
    factions = active_factions if active_factions is not None else [0 for _ in candidate_counts]
    for env_index, count in enumerate(candidate_counts):
        candidates = [[float(slot == feature), 0.5] for slot in range(count) for feature in [slot]]
        decisions.append(
            decision(
                env_index,
                observation=[float(env_index), 1.0, 0.0],
                candidates=candidates,
                reward=rewards[env_index],
                done=dones[env_index],
                active_faction=factions[env_index],
            )
        )
    return batch_to_arrays(
        EnvBatch(
            decisions=tuple(decisions),
            observation_length=3,
            action_length=2,
        )
    )
