from __future__ import annotations

import time

import numpy as np

from rootbuddy_rl import Faction, RootBuddyVecEnv, VecEnvArrays, VecEnvConfig, sample_random_actions


def test_random_agent_drives_go_env_end_to_end() -> None:
    config = VecEnvConfig(
        num_envs=2,
        base_seed=909,
        max_steps=32,
        factions=[Faction.MARQUISE, Faction.EYRIE],
        player_faction=Faction.MARQUISE,
        track_all_hands=True,
    )
    rng = np.random.default_rng(909)

    completed_episodes = 0
    applied_steps = 0
    started = time.perf_counter()

    with RootBuddyVecEnv(config) as env:
        batch = env.reset()
        assert_batch_shape(batch, num_envs=2)

        for _ in range(128):
            actions = sample_random_actions(batch, rng)
            assert np.all(actions >= 0)

            batch = env.step(actions.tolist())
            applied_steps += batch.num_envs
            assert_batch_shape(batch, num_envs=2)

            if np.any(batch.dones):
                completed_episodes += int(np.count_nonzero(batch.dones))
                if completed_episodes >= 2:
                    break
                batch = env.reset()
                assert_batch_shape(batch, num_envs=2)

    elapsed = time.perf_counter() - started
    assert completed_episodes >= 2
    assert applied_steps > 0
    assert applied_steps / elapsed > 0


def assert_batch_shape(batch: VecEnvArrays, *, num_envs: int) -> None:
    assert batch.num_envs == num_envs
    assert batch.observations.shape == (num_envs, batch.observation_length)
    assert batch.candidate_actions.shape[0] == num_envs
    assert batch.candidate_actions.shape[2] == batch.action_length
    assert batch.action_mask.shape == batch.candidate_actions.shape[:2]
    assert batch.rewards.shape == (num_envs,)
    assert batch.dones.shape == (num_envs,)
    assert batch.active_factions.shape == (num_envs,)
    assert batch.candidate_counts.shape == (num_envs,)
    np.testing.assert_array_equal(batch.action_mask.sum(axis=1), batch.candidate_counts)
