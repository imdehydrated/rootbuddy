from __future__ import annotations

from pathlib import Path

import torch

from rootbuddy_rl.engine_client import EnvBatch, Faction, VecEnvConfig
from rootbuddy_rl.model import CandidatePolicyValueNet
from rootbuddy_rl.train import (
    TrainConfig,
    build_model_from_batch,
    parse_args,
    terminal_outcomes,
    train,
    two_player_train_config,
)
from rootbuddy_rl.vec_env import VecEnvArrays, batch_to_arrays
from test_vec_env import decision


class FakeTrainingEnv:
    def __init__(self) -> None:
        self.reset_calls = 0
        self.step_actions: list[list[int]] = []

    def reset(self) -> VecEnvArrays:
        self.reset_calls += 1
        return make_batch(
            rewards=[0.0, 0.0],
            dones=[False, False],
            candidate_counts=[2, 2],
        )

    def step(self, action_indices: list[int]) -> VecEnvArrays:
        self.step_actions.append(list(action_indices))
        done = len(self.step_actions) >= 2
        return make_batch(
            rewards=[1.0, 0.0],
            dones=[False, done],
            candidate_counts=[2, 2],
            winner=2 if done else 0,
            victory_points=(11, 7, 30, 5) if done else (1, 0, 0, 0),
            step=len(self.step_actions),
        )


def test_build_model_from_batch_uses_env_dimensions() -> None:
    config = TrainConfig(
        env_config=VecEnvConfig(num_envs=1, base_seed=707),
        updates=1,
        rollout_steps=1,
        torso_hidden_dims=(8,),
        head_hidden_dim=8,
    )
    model = build_model_from_batch(make_batch([0.0], [False], [3]), config)

    output = model(
        torch.zeros((1, 3)),
        torch.zeros((1, 3, 2)),
        torch.ones((1, 3), dtype=torch.bool),
        torch.zeros((1,), dtype=torch.long),
    )

    assert output.logits.shape == (1, 3)
    assert output.values.shape == (1,)


def test_train_runs_update_and_writes_checkpoint(tmp_path: Path) -> None:
    torch.manual_seed(707)
    env = FakeTrainingEnv()
    model = CandidatePolicyValueNet(
        observation_dim=3,
        action_dim=2,
        torso_hidden_dims=(8,),
        head_hidden_dim=8,
    )
    before = [parameter.detach().clone() for parameter in model.parameters()]
    config = TrainConfig(
        env_config=VecEnvConfig(num_envs=2, base_seed=707),
        updates=1,
        rollout_steps=2,
        ppo_epochs=1,
        minibatch_size=2,
        checkpoint_dir=tmp_path,
        checkpoint_every=1,
        torso_hidden_dims=(8,),
        head_hidden_dim=8,
        seed=707,
    )

    result = train(config, env=env, model=model)  # type: ignore[arg-type]

    assert env.reset_calls == 2
    assert len(env.step_actions) == 2
    assert len(result.metrics) == 1
    metrics = result.metrics[0]
    assert metrics.completed_episodes == 1
    assert metrics.mean_game_length == 2.0
    assert metrics.per_faction_win_rate == (0.0, 0.0, 1.0, 0.0)
    assert metrics.per_faction_terminal_vp == (11.0, 7.0, 30.0, 5.0)
    assert metrics.checkpoint_path is not None
    assert Path(metrics.checkpoint_path).exists()
    assert any(
        not torch.allclose(previous, current)
        for previous, current in zip(before, result.model.parameters(), strict=True)
    )


def test_terminal_outcomes_returns_zeroes_without_completed_games() -> None:
    outcomes = terminal_outcomes(make_batch([0.0], [False], [1]))

    assert outcomes["mean_game_length"] == 0.0
    assert outcomes["per_faction_win_rate"] == (0.0, 0.0, 0.0, 0.0)
    assert outcomes["per_faction_terminal_vp"] == (0.0, 0.0, 0.0, 0.0)


def test_terminal_outcomes_does_not_credit_truncated_games_as_wins() -> None:
    outcomes = terminal_outcomes(
        make_batch(
            rewards=[0.0],
            dones=[True],
            candidate_counts=[1],
            truncations=[True],
            winner=0,
            victory_points=(4, 3, 2, 1),
            step=32,
        )
    )

    assert outcomes["mean_game_length"] == 32.0
    assert outcomes["per_faction_win_rate"] == (0.0, 0.0, 0.0, 0.0)
    assert outcomes["per_faction_terminal_vp"] == (4.0, 3.0, 2.0, 1.0)


def test_cli_args_build_two_player_training_config() -> None:
    args = parse_args(
        [
            "--updates",
            "3",
            "--num-envs",
            "2",
            "--base-seed",
            "909",
            "--rollout-steps",
            "5",
            "--track-all-hands",
        ]
    )

    config = two_player_train_config(args)

    assert config.updates == 3
    assert config.rollout_steps == 5
    assert config.env_config.num_envs == 2
    assert config.env_config.base_seed == 909
    assert config.env_config.factions == [Faction.MARQUISE, Faction.EYRIE]
    assert config.env_config.track_all_hands


def make_batch(
    rewards: list[float],
    dones: list[bool],
    candidate_counts: list[int],
    *,
    truncations: list[bool] | None = None,
    winner: int = 0,
    victory_points: tuple[int, ...] = (0, 0, 0, 0),
    step: int = 0,
) -> VecEnvArrays:
    decisions = []
    truncated_rows = truncations if truncations is not None else [False for _ in dones]
    for env_index, count in enumerate(candidate_counts):
        candidates = [[float(slot == 0), float(slot == 1)] for slot in range(count)]
        decisions.append(
            decision(
                env_index,
                observation=[float(env_index), 1.0, 0.0],
                candidates=candidates,
                step=step,
                reward=rewards[env_index],
                done=dones[env_index],
                truncated=truncated_rows[env_index],
                winner=winner,
                victory_points=victory_points,
            )
        )
    return batch_to_arrays(
        EnvBatch(
            decisions=tuple(decisions),
            observation_length=3,
            action_length=2,
        )
    )
