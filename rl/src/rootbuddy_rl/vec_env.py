"""Gym-style vector environment adapter over :mod:`rootbuddy_rl.engine_client`."""

from __future__ import annotations

from dataclasses import dataclass
from typing import Iterable

import numpy as np

from .engine_client import EngineClient, EnvBatch, EnvDecision, VecEnvConfig


@dataclass(frozen=True)
class VecEnvArrays:
    observations: np.ndarray
    candidate_actions: np.ndarray
    candidate_rewards: np.ndarray
    action_mask: np.ndarray
    candidate_counts: np.ndarray
    rewards: np.ndarray
    dones: np.ndarray
    truncations: np.ndarray
    active_factions: np.ndarray
    current_phases: np.ndarray
    current_steps: np.ndarray
    episodes: np.ndarray
    steps: np.ndarray
    winners: np.ndarray
    victory_points: np.ndarray
    decisions: tuple[EnvDecision, ...]

    @property
    def num_envs(self) -> int:
        return int(self.observations.shape[0])

    @property
    def max_candidates(self) -> int:
        return int(self.candidate_actions.shape[1])

    @property
    def observation_length(self) -> int:
        return int(self.observations.shape[1])

    @property
    def action_length(self) -> int:
        return int(self.candidate_actions.shape[2])


class RootBuddyVecEnv:
    def __init__(
        self,
        config: VecEnvConfig | dict[str, object],
        *,
        client: EngineClient | None = None,
    ) -> None:
        self.config = config
        self.client = client if client is not None else EngineClient()
        self._owns_client = client is None
        self._configured = False
        self.last_batch: VecEnvArrays | None = None

    def __enter__(self) -> RootBuddyVecEnv:
        self.configure()
        return self

    def __exit__(self, exc_type: object, exc: object, traceback: object) -> None:
        self.close()

    def configure(self) -> None:
        if self._configured:
            return
        self.client.configure(self.config)
        self._configured = True

    def reset(self, env_indices: Iterable[int] | None = None) -> VecEnvArrays:
        self.configure()
        reset_batch = batch_to_arrays(self.client.reset(env_indices))
        if env_indices is None or self.last_batch is None:
            self.last_batch = reset_batch
        else:
            self.last_batch = merge_batches(self.last_batch, reset_batch)
        return self.last_batch

    def step(self, action_indices: Iterable[int]) -> VecEnvArrays:
        self.configure()
        self.last_batch = batch_to_arrays(self.client.step(action_indices))
        return self.last_batch

    def close(self) -> None:
        if self._owns_client:
            self.client.close()


def batch_to_arrays(batch: EnvBatch) -> VecEnvArrays:
    decisions = batch.decisions
    num_envs = len(decisions)
    observation_length = batch.observation_length
    action_length = batch.action_length
    max_candidates = max((decision.candidate_count for decision in decisions), default=0)

    observations = np.zeros((num_envs, observation_length), dtype=np.float32)
    candidate_actions = np.zeros((num_envs, max_candidates, action_length), dtype=np.float32)
    candidate_rewards = np.zeros((num_envs, max_candidates), dtype=np.float32)
    action_mask = np.zeros((num_envs, max_candidates), dtype=np.bool_)
    candidate_counts = np.zeros((num_envs,), dtype=np.int64)
    rewards = np.zeros((num_envs,), dtype=np.float32)
    dones = np.zeros((num_envs,), dtype=np.bool_)
    truncations = np.zeros((num_envs,), dtype=np.bool_)
    active_factions = np.zeros((num_envs,), dtype=np.int64)
    current_phases = np.zeros((num_envs,), dtype=np.int64)
    current_steps = np.zeros((num_envs,), dtype=np.int64)
    episodes = np.zeros((num_envs,), dtype=np.int64)
    steps = np.zeros((num_envs,), dtype=np.int64)
    winners = np.zeros((num_envs,), dtype=np.int64)
    victory_points = np.zeros((num_envs, 4), dtype=np.int64)

    for row, decision in enumerate(decisions):
        observations[row, :] = _fit_vector(decision.observation, observation_length)
        count = decision.candidate_count
        candidate_counts[row] = count
        rewards[row] = decision.reward
        dones[row] = decision.done
        truncations[row] = decision.truncated
        active_factions[row] = decision.active_faction
        current_phases[row] = decision.current_phase
        current_steps[row] = decision.current_step
        episodes[row] = decision.episode
        steps[row] = decision.step
        winners[row] = decision.winner
        point_count = min(len(decision.victory_points), victory_points.shape[1])
        if point_count > 0:
            victory_points[row, :point_count] = decision.victory_points[:point_count]
        if count > 0:
            action_mask[row, :count] = True
            candidate_rewards[row, :count] = _fit_vector(decision.candidate_rewards, count)
            candidate_actions[row, :count, :] = _fit_matrix(
                decision.candidate_actions,
                rows=count,
                cols=action_length,
            )

    return VecEnvArrays(
        observations=observations,
        candidate_actions=candidate_actions,
        candidate_rewards=candidate_rewards,
        action_mask=action_mask,
        candidate_counts=candidate_counts,
        rewards=rewards,
        dones=dones,
        truncations=truncations,
        active_factions=active_factions,
        current_phases=current_phases,
        current_steps=current_steps,
        episodes=episodes,
        steps=steps,
        winners=winners,
        victory_points=victory_points,
        decisions=decisions,
    )


def merge_batches(current: VecEnvArrays, reset: VecEnvArrays) -> VecEnvArrays:
    decisions = list(current.decisions)
    index_by_env = {decision.env_index: row for row, decision in enumerate(decisions)}
    for decision in reset.decisions:
        row = index_by_env.get(decision.env_index)
        if row is None:
            raise ValueError(f"reset returned unknown env index {decision.env_index}")
        decisions[row] = decision
    return batch_to_arrays(
        EnvBatch(
            decisions=tuple(decisions),
            observation_length=current.observation_length,
            action_length=current.action_length,
        )
    )


def sample_random_actions(batch: VecEnvArrays, rng: np.random.Generator | None = None) -> np.ndarray:
    generator = rng if rng is not None else np.random.default_rng()
    actions = np.full((batch.num_envs,), -1, dtype=np.int64)
    for env_index, count in enumerate(batch.candidate_counts):
        if count > 0:
            actions[env_index] = int(generator.integers(0, int(count)))
    return actions


def _fit_vector(values: np.ndarray, length: int) -> np.ndarray:
    fitted = np.zeros((length,), dtype=np.float32)
    size = min(values.size, length)
    if size > 0:
        fitted[:size] = values[:size]
    return fitted


def _fit_matrix(values: np.ndarray, *, rows: int, cols: int) -> np.ndarray:
    fitted = np.zeros((rows, cols), dtype=np.float32)
    if values.size == 0:
        return fitted
    matrix = np.asarray(values, dtype=np.float32).reshape((-1, cols))
    row_count = min(matrix.shape[0], rows)
    fitted[:row_count, :] = matrix[:row_count, :]
    return fitted
