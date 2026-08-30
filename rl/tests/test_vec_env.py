from __future__ import annotations

import numpy as np

from rootbuddy_rl.engine_client import ConfigResponse, EnvBatch, EnvDecision, EyrieDiagnostics, VecEnvConfig
from rootbuddy_rl.vec_env import RootBuddyVecEnv, batch_to_arrays, sample_random_actions


class FakeClient:
    def __init__(self, reset_batch: EnvBatch, step_batch: EnvBatch) -> None:
        self.reset_batch = reset_batch
        self.step_batch = step_batch
        self.configured_with: object | None = None
        self.step_actions: list[list[int]] = []
        self.closed = False

    def configure(self, config: object) -> ConfigResponse:
        self.configured_with = config
        return ConfigResponse(config={}, observation_length=2, action_length=3)

    def reset(self, env_indices: object | None = None) -> EnvBatch:
        if env_indices is None:
            return self.reset_batch
        return self.step_batch

    def step(self, action_indices: object) -> EnvBatch:
        self.step_actions.append(list(action_indices))
        return self.step_batch

    def close(self) -> None:
        self.closed = True


def test_batch_to_arrays_pads_candidates_and_builds_mask() -> None:
    batch = EnvBatch(
        decisions=(
            decision(
                0,
                observation=[1.0, 2.0],
                candidates=[[1.0, 0.0, 0.0], [0.0, 1.0, 0.0]],
                candidate_rewards=[0.5, 1.5],
            ),
            decision(
                1,
                observation=[3.0, 4.0],
                candidates=[[0.0, 0.0, 1.0]],
                candidate_rewards=[2.5],
                reward=2.0,
                done=True,
                eyrie_roosts=2,
                eyrie_decree_counts=(1, 2, 3, 4),
            ),
        ),
        observation_length=2,
        action_length=3,
    )

    arrays = batch_to_arrays(batch)

    assert arrays.observations.shape == (2, 2)
    assert arrays.candidate_actions.shape == (2, 2, 3)
    assert arrays.action_mask.tolist() == [[True, True], [True, False]]
    assert arrays.candidate_counts.tolist() == [2, 1]
    assert arrays.rewards.tolist() == [0.0, 2.0]
    assert arrays.dones.tolist() == [False, True]
    assert arrays.truncations.tolist() == [False, False]
    assert arrays.current_phases.tolist() == [0, 0]
    assert arrays.current_steps.tolist() == [0, 0]
    assert arrays.round_numbers.tolist() == [0, 0]
    assert arrays.eyrie_roosts.tolist() == [0, 2]
    assert arrays.eyrie_decree_counts.tolist() == [[0, 0, 0, 0], [1, 2, 3, 4]]
    assert arrays.candidate_rewards.tolist() == [[0.5, 1.5], [2.5, 0.0]]
    assert arrays.victory_points.tolist() == [[0, 0, 0, 0], [0, 0, 0, 0]]
    np.testing.assert_array_equal(arrays.candidate_actions[1, 1], np.zeros((3,), dtype=np.float32))


def test_rootbuddy_vec_env_configures_once_and_delegates_step() -> None:
    reset_batch = EnvBatch(
        decisions=(decision(0, observation=[1.0, 0.0], candidates=[[1.0, 0.0, 0.0]]),),
        observation_length=2,
        action_length=3,
    )
    step_batch = EnvBatch(
        decisions=(decision(0, observation=[0.0, 1.0], candidates=[[0.0, 1.0, 0.0]], step=1, reward=1.0),),
        observation_length=2,
        action_length=3,
    )
    client = FakeClient(reset_batch=reset_batch, step_batch=step_batch)
    config = VecEnvConfig(num_envs=1, base_seed=707)
    env = RootBuddyVecEnv(config, client=client)  # type: ignore[arg-type]

    reset = env.reset()
    stepped = env.step([0])

    assert client.configured_with is config
    assert client.step_actions == [[0]]
    assert reset.steps.tolist() == [0]
    assert stepped.steps.tolist() == [1]
    assert stepped.rewards.tolist() == [1.0]
    assert stepped.truncations.tolist() == [False]
    assert stepped.victory_points.tolist() == [[0, 0, 0, 0]]
    assert env.last_batch is stepped


def test_rootbuddy_vec_env_merges_partial_reset_by_env_index() -> None:
    reset_batch = EnvBatch(
        decisions=(
            decision(0, observation=[1.0], candidates=[[1.0]], step=0),
            decision(1, observation=[2.0], candidates=[[2.0]], step=0),
        ),
        observation_length=1,
        action_length=1,
    )
    partial_reset_batch = EnvBatch(
        decisions=(decision(1, observation=[3.0], candidates=[[3.0], [4.0]], step=0),),
        observation_length=1,
        action_length=1,
    )
    client = FakeClient(reset_batch=reset_batch, step_batch=partial_reset_batch)
    env = RootBuddyVecEnv(VecEnvConfig(num_envs=2, base_seed=707), client=client)  # type: ignore[arg-type]

    initial = env.reset()
    merged = env.reset(env_indices=[1])

    assert initial.observations[:, 0].tolist() == [1.0, 2.0]
    assert merged.observations[:, 0].tolist() == [1.0, 3.0]
    assert merged.candidate_counts.tolist() == [1, 2]


def test_sample_random_actions_respects_candidate_counts() -> None:
    arrays = batch_to_arrays(
        EnvBatch(
            decisions=(
                decision(0, observation=[0.0], candidates=[]),
                decision(1, observation=[0.0], candidates=[[1.0], [2.0], [3.0]]),
            ),
            observation_length=1,
            action_length=1,
        )
    )

    actions = sample_random_actions(arrays, np.random.default_rng(123))

    assert actions[0] == -1
    assert 0 <= actions[1] < 3


def decision(
    env_index: int,
    *,
    observation: list[float],
    candidates: list[list[float]],
    candidate_rewards: list[float] | None = None,
    step: int = 0,
    reward: float = 0.0,
    done: bool = False,
    truncated: bool = False,
    winner: int = 0,
    active_faction: int = 0,
    current_phase: int = 0,
    current_step: int = 0,
    round_number: int = 0,
    eyrie_roosts: int = 0,
    eyrie_warrior_supply: int = 0,
    eyrie_decree_counts: tuple[int, int, int, int] = (0, 0, 0, 0),
    eyrie_current_decree_column: int = -1,
    eyrie_decree_columns_resolved: int = 0,
    eyrie_decree_cards_resolved: int = 0,
    eyrie_cards_added_to_decree: int = 0,
    victory_points: tuple[int, ...] = (0, 0, 0, 0),
    legal_action_types: tuple[int, ...] | None = None,
) -> EnvDecision:
    action_length = len(candidates[0]) if candidates else 1
    return EnvDecision(
        env_index=env_index,
        episode=0,
        step=step,
        round_number=round_number,
        active_faction=active_faction,
        current_phase=current_phase,
        current_step=current_step,
        eyrie=EyrieDiagnostics(
            roosts_placed=eyrie_roosts,
            warrior_supply=eyrie_warrior_supply,
            decree_column_counts=eyrie_decree_counts,
            current_decree_column=eyrie_current_decree_column,
            decree_columns_resolved=eyrie_decree_columns_resolved,
            decree_cards_resolved=eyrie_decree_cards_resolved,
            cards_added_to_decree=eyrie_cards_added_to_decree,
        ),
        observation=np.asarray(observation, dtype=np.float32),
        candidate_actions=np.asarray(candidates, dtype=np.float32).reshape((len(candidates), action_length)),
        candidate_rewards=np.asarray(
            candidate_rewards if candidate_rewards is not None else [0.0 for _ in candidates],
            dtype=np.float32,
        ),
        candidate_count=len(candidates),
        reward=reward,
        done=done,
        truncated=truncated,
        winner=winner,
        winning_coalition=(),
        victory_points=victory_points,
        legal_action_types=legal_action_types if legal_action_types is not None else tuple(range(len(candidates))),
        observation_length=len(observation),
        action_length=action_length,
    )
