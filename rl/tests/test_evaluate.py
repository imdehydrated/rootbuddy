from __future__ import annotations

from pathlib import Path

import numpy as np
import torch

from rootbuddy_rl.engine_client import EnvBatch, VecEnvConfig
from rootbuddy_rl.evaluate import (
    EvaluationConfig,
    GreedyVPAgent,
    ModelAgent,
    evaluate_agents,
    load_model_from_checkpoint,
    parse_args,
    select_actions_by_faction,
    two_player_eval_config,
    format_metrics,
)
from rootbuddy_rl.model import CandidatePolicyValueNet
from rootbuddy_rl.vec_env import VecEnvArrays, batch_to_arrays
from test_vec_env import decision


class ConstantAgent:
    def __init__(self, action: int) -> None:
        self.action = action

    def select_actions(self, batch: VecEnvArrays) -> np.ndarray:
        return np.full((batch.num_envs,), self.action, dtype=np.int64)


class FakeEvalEnv:
    def __init__(self) -> None:
        self.reset_indices: list[list[int] | None] = []
        self.step_actions: list[list[int]] = []

    def reset(self, env_indices: list[int] | None = None) -> VecEnvArrays:
        self.reset_indices.append(env_indices)
        if env_indices is None:
            return make_batch(
                rewards=[0.0, 0.0],
                dones=[False, False],
                active_factions=[0, 2],
                candidate_rewards=[[0.0, 2.0], [3.0, 1.0]],
                legal_action_types=[[10, 11], [7, 21]],
                victory_points=[(0, 0, 0, 0), (0, 5, 0, 0)],
                eyrie_roosts=[1, 1],
                eyrie_decree_counts=[(1, 1, 0, 0), (1, 1, 0, 0)],
            )
        return make_batch(
            rewards=[0.0, 0.0],
            dones=[False, False],
            active_factions=[0, 2],
            candidate_rewards=[[1.0, 0.0], [2.0, 3.0]],
            legal_action_types=[[12, 13], [22, 23]],
            victory_points=[(0, 0, 0, 0), (4, 6, 8, 0)],
            eyrie_roosts=[1, 2],
            eyrie_decree_counts=[(1, 1, 1, 0), (1, 1, 1, 0)],
        )

    def step(self, action_indices: list[int]) -> VecEnvArrays:
        self.step_actions.append(list(action_indices))
        if len(self.step_actions) == 1:
            return make_batch(
                rewards=[2.0, 0.5],
                dones=[True, False],
                active_factions=[0, 2],
                candidate_rewards=[[0.0], [2.0, 3.0]],
                legal_action_types=[[14], [24, 25]],
                winners=[0, 0],
                victory_points=[(30, 7, 5, 0), (4, 8, 6, 0)],
                steps=[4, 4],
                round_numbers=[2, 2],
                current_phases=[2, 1],
                current_steps=[4, 3],
                eyrie_roosts=[1, 2],
                eyrie_decree_counts=[(1, 1, 0, 0), (1, 1, 1, 0)],
            )
        return make_batch(
            rewards=[0.0, 3.0],
            dones=[False, True],
            active_factions=[0, 2],
            candidate_rewards=[[1.0], [1.0]],
            legal_action_types=[[15], [26]],
            winners=[0, 2],
            victory_points=[(1, 0, 0, 0), (6, 30, 8, 0)],
            steps=[1, 5],
            round_numbers=[1, 3],
            current_phases=[0, 2],
            current_steps=[1, 4],
            eyrie_roosts=[0, 2],
            eyrie_decree_counts=[(1, 1, 1, 0), (1, 1, 1, 0)],
        )


def test_greedy_vp_agent_picks_highest_candidate_reward() -> None:
    batch = make_batch(
        rewards=[0.0, 0.0],
        dones=[False, False],
        active_factions=[0, 2],
        candidate_rewards=[[0.0, 2.5, 1.0], [-1.0, 0.5]],
    )

    actions = GreedyVPAgent().select_actions(batch)

    assert actions.tolist() == [1, 1]


def test_select_actions_by_faction_routes_to_matching_agent() -> None:
    batch = make_batch(
        rewards=[0.0, 0.0],
        dones=[False, False],
        active_factions=[0, 2],
        candidate_rewards=[[0.0, 1.0], [0.0, 1.0]],
    )

    actions = select_actions_by_faction(
        batch,
        agents={0: ConstantAgent(0)},
        default_agent=ConstantAgent(1),
    )

    assert actions.tolist() == [0, 1]


def test_evaluate_agents_counts_completed_games_and_partial_resets() -> None:
    env = FakeEvalEnv()

    metrics = evaluate_agents(
        EvaluationConfig(env_config=VecEnvConfig(num_envs=2, base_seed=707), episodes=2),
        agents={},
        default_agent=GreedyVPAgent(),
        env=env,  # type: ignore[arg-type]
    )

    assert env.reset_indices == [None, [0]]
    assert env.step_actions == [[1, 0], [0, 1]]
    assert metrics.episodes == 2
    assert metrics.transitions == 4
    assert metrics.truncated_episodes == 0
    assert metrics.mean_reward == 5.5 / 4
    assert metrics.mean_game_length == 4.5
    assert metrics.per_faction_win_rate == (0.5, 0.0, 0.5, 0.0)
    assert metrics.per_faction_terminal_vp == (18.0, 18.5, 6.5, 0.0)
    assert metrics.diagnostics.terminal_reasons == (("win", 2),)
    assert metrics.diagnostics.final_active_factions == ((0, 1), (2, 1))
    assert metrics.diagnostics.final_phase_steps == (((2, 4), 2),)
    assert metrics.diagnostics.action_type_counts == ((7, 1), (11, 1), (12, 1), (23, 1))
    assert metrics.diagnostics.per_faction_action_counts == (2, 0, 2, 0)
    assert metrics.diagnostics.mean_candidate_count == 2.0
    assert metrics.diagnostics.mean_terminal_round_number == 2.5
    assert metrics.diagnostics.eyrie_action_type_counts == ((7, 1), (23, 1))
    assert metrics.diagnostics.eyrie_score_roosts_count == 1
    assert metrics.diagnostics.eyrie_score_roosts_vp == 24
    assert metrics.diagnostics.eyrie_roost_builds == 1
    assert metrics.diagnostics.eyrie_roost_losses == 1
    assert metrics.diagnostics.eyrie_roost_losses_on_marquise_turn == 1
    assert metrics.diagnostics.eyrie_cards_added_by_decree_column == (0, 0, 1, 0)
    assert metrics.diagnostics.mean_eyrie_roosts == 1.25
    assert metrics.diagnostics.mean_terminal_eyrie_roosts == 1.5
    assert metrics.diagnostics.mean_terminal_eyrie_vp == 18.5
    assert metrics.diagnostics.mean_terminal_eyrie_decree_counts == (1.0, 1.0, 0.5, 0.0)
    formatted = format_metrics(metrics)
    assert "decision_steps=4.50" in formatted
    assert "game_length=" not in formatted
    assert "'mean_terminal_round_number': 2.5" in formatted
    assert "'eyrie_score_roosts_vp': 24" in formatted


def test_evaluate_agents_records_truncation_diagnostics() -> None:
    env = FakeTruncatingEvalEnv()

    metrics = evaluate_agents(
        EvaluationConfig(env_config=VecEnvConfig(num_envs=1, base_seed=707, max_steps=4), episodes=1),
        agents={},
        default_agent=ConstantAgent(0),
        env=env,  # type: ignore[arg-type]
    )

    assert metrics.truncated_episodes == 1
    assert metrics.diagnostics.terminal_reasons == (("max_steps", 1),)
    assert metrics.diagnostics.final_active_factions == ((2, 1),)
    assert metrics.diagnostics.final_phase_steps == (((1, 3), 1),)
    assert metrics.diagnostics.action_type_counts == ((30, 1), (31, 1))
    assert metrics.diagnostics.truncated_recent_action_type_counts == ((30, 1), (31, 1))


def test_evaluate_agents_records_eyrie_turmoil_diagnostics() -> None:
    env = FakeEyrieTurmoilEvalEnv()

    metrics = evaluate_agents(
        EvaluationConfig(env_config=VecEnvConfig(num_envs=1, base_seed=707), episodes=1),
        agents={},
        default_agent=ConstantAgent(0),
        env=env,  # type: ignore[arg-type]
    )

    assert metrics.diagnostics.eyrie_action_type_counts == ((18, 1),)
    assert metrics.diagnostics.eyrie_turmoil_count == 1
    assert metrics.diagnostics.eyrie_turmoil_vp_lost == 2
    assert metrics.diagnostics.eyrie_turmoil_by_decree_column == ((3, 1),)


class FakeTruncatingEvalEnv:
    def __init__(self) -> None:
        self.step_actions: list[list[int]] = []

    def reset(self, env_indices: list[int] | None = None) -> VecEnvArrays:
        return make_batch(
            rewards=[0.0],
            dones=[False],
            active_factions=[0],
            candidate_rewards=[[0.0]],
            legal_action_types=[[30]],
        )

    def step(self, action_indices: list[int]) -> VecEnvArrays:
        self.step_actions.append(list(action_indices))
        if len(self.step_actions) == 1:
            return make_batch(
                rewards=[0.0],
                dones=[False],
                active_factions=[2],
                candidate_rewards=[[0.0]],
                legal_action_types=[[31]],
                steps=[1],
                round_numbers=[1],
                current_phases=[0],
                current_steps=[1],
            )
        return make_batch(
            rewards=[0.0],
            dones=[True],
            active_factions=[2],
            candidate_rewards=[[]],
            legal_action_types=[[]],
            truncations=[True],
            steps=[4],
            round_numbers=[2],
            current_phases=[1],
            current_steps=[3],
        )


class FakeEyrieTurmoilEvalEnv:
    def reset(self, env_indices: list[int] | None = None) -> VecEnvArrays:
        return make_batch(
            rewards=[0.0],
            dones=[False],
            active_factions=[2],
            candidate_rewards=[[0.0]],
            legal_action_types=[[18]],
            victory_points=[(0, 5, 0, 0)],
            eyrie_roosts=[1],
            eyrie_decree_counts=[(1, 1, 1, 1)],
            eyrie_current_decree_columns=[3],
        )

    def step(self, action_indices: list[int]) -> VecEnvArrays:
        return make_batch(
            rewards=[-2.0],
            dones=[True],
            active_factions=[2],
            candidate_rewards=[[]],
            legal_action_types=[[]],
            winners=[0],
            victory_points=[(30, 3, 0, 0)],
            eyrie_roosts=[1],
            eyrie_decree_counts=[(1, 1, 0, 0)],
        )


def test_two_player_eval_config_tracks_all_hands_by_default() -> None:
    config = two_player_eval_config(parse_args([]))

    assert config.env_config.track_all_hands
    assert config.env_config.step_penalty == 0.01
    assert config.env_config.truncation_penalty == 10.0
    assert not config.env_config.disable_eyrie_reward_shaping
    assert config.env_config.eyrie_turmoil_penalty == 1.0
    assert config.env_config.eyrie_turmoil_vp_loss_penalty == 0.5
    assert config.env_config.eyrie_score_roosts_bonus == 0.25
    assert config.env_config.eyrie_roost_build_bonus == 0.5
    assert config.env_config.eyrie_roost_loss_penalty == 0.75
    assert config.env_config.eyrie_build_decree_card_penalty == 0.05


def test_model_agent_restores_training_mode() -> None:
    model = CandidatePolicyValueNet(
        observation_dim=3,
        action_dim=2,
        torso_hidden_dims=(8,),
        head_hidden_dim=8,
    )
    model.train()
    batch = make_batch(
        rewards=[0.0],
        dones=[False],
        active_factions=[0],
        candidate_rewards=[[0.0, 1.0]],
    )

    actions = ModelAgent(model, deterministic=True).select_actions(batch)

    assert actions.shape == (1,)
    assert model.training


def test_load_model_from_checkpoint_uses_saved_architecture(tmp_path: Path) -> None:
    model = CandidatePolicyValueNet(
        observation_dim=3,
        action_dim=2,
        torso_hidden_dims=(8,),
        head_hidden_dim=8,
    )
    path = tmp_path / "checkpoint.pt"
    torch.save(
        {
            "config": {"torso_hidden_dims": (8,), "head_hidden_dim": 8},
            "model_state_dict": model.state_dict(),
        },
        path,
    )

    loaded = load_model_from_checkpoint(path, observation_dim=3, action_dim=2)

    assert loaded.observation_dim == 3
    assert loaded.action_dim == 2
    assert not loaded.training


def make_batch(
    *,
    rewards: list[float],
    dones: list[bool],
    active_factions: list[int],
    candidate_rewards: list[list[float]],
    legal_action_types: list[list[int]] | None = None,
    winners: list[int] | None = None,
    truncations: list[bool] | None = None,
    victory_points: list[tuple[int, ...]] | None = None,
    eyrie_victory_points: list[int] | None = None,
    steps: list[int] | None = None,
    round_numbers: list[int] | None = None,
    eyrie_roosts: list[int] | None = None,
    eyrie_warrior_supply: list[int] | None = None,
    eyrie_decree_counts: list[tuple[int, int, int, int]] | None = None,
    eyrie_current_decree_columns: list[int] | None = None,
    eyrie_decree_columns_resolved: list[int] | None = None,
    eyrie_decree_cards_resolved: list[int] | None = None,
    eyrie_cards_added_to_decree: list[int] | None = None,
    current_phases: list[int] | None = None,
    current_steps: list[int] | None = None,
) -> VecEnvArrays:
    decisions = []
    winner_rows = winners if winners is not None else [0 for _ in rewards]
    truncated_rows = truncations if truncations is not None else [False for _ in rewards]
    vp_rows = victory_points if victory_points is not None else [(0, 0, 0, 0) for _ in rewards]
    eyrie_vp_rows = eyrie_victory_points if eyrie_victory_points is not None else [
        row[1] if len(row) > 1 else 0 for row in vp_rows
    ]
    step_rows = steps if steps is not None else [0 for _ in rewards]
    round_rows = round_numbers if round_numbers is not None else [0 for _ in rewards]
    eyrie_roost_rows = eyrie_roosts if eyrie_roosts is not None else [0 for _ in rewards]
    eyrie_supply_rows = eyrie_warrior_supply if eyrie_warrior_supply is not None else [0 for _ in rewards]
    eyrie_decree_rows = eyrie_decree_counts if eyrie_decree_counts is not None else [
        (0, 0, 0, 0) for _ in rewards
    ]
    eyrie_column_rows = eyrie_current_decree_columns if eyrie_current_decree_columns is not None else [
        -1 for _ in rewards
    ]
    eyrie_columns_resolved_rows = eyrie_decree_columns_resolved if eyrie_decree_columns_resolved is not None else [
        0 for _ in rewards
    ]
    eyrie_cards_resolved_rows = eyrie_decree_cards_resolved if eyrie_decree_cards_resolved is not None else [
        0 for _ in rewards
    ]
    eyrie_cards_added_rows = eyrie_cards_added_to_decree if eyrie_cards_added_to_decree is not None else [
        0 for _ in rewards
    ]
    phase_rows = current_phases if current_phases is not None else [0 for _ in rewards]
    turn_step_rows = current_steps if current_steps is not None else [0 for _ in rewards]
    action_type_rows = legal_action_types if legal_action_types is not None else [
        list(range(len(row))) for row in candidate_rewards
    ]
    for env_index, rewards_row in enumerate(candidate_rewards):
        candidates = [[float(index == 0), float(index == 1)] for index, _ in enumerate(rewards_row)]
        decisions.append(
            decision(
                env_index,
                observation=[float(env_index), 1.0, 0.0],
                candidates=candidates,
                candidate_rewards=rewards_row,
                step=step_rows[env_index],
                round_number=round_rows[env_index],
                eyrie_victory_points=eyrie_vp_rows[env_index],
                eyrie_roosts=eyrie_roost_rows[env_index],
                eyrie_warrior_supply=eyrie_supply_rows[env_index],
                eyrie_decree_counts=eyrie_decree_rows[env_index],
                eyrie_current_decree_column=eyrie_column_rows[env_index],
                eyrie_decree_columns_resolved=eyrie_columns_resolved_rows[env_index],
                eyrie_decree_cards_resolved=eyrie_cards_resolved_rows[env_index],
                eyrie_cards_added_to_decree=eyrie_cards_added_rows[env_index],
                reward=rewards[env_index],
                done=dones[env_index],
                truncated=truncated_rows[env_index],
                winner=winner_rows[env_index],
                active_faction=active_factions[env_index],
                current_phase=phase_rows[env_index],
                current_step=turn_step_rows[env_index],
                victory_points=vp_rows[env_index],
                legal_action_types=tuple(action_type_rows[env_index]),
            )
        )
    return batch_to_arrays(
        EnvBatch(
            decisions=tuple(decisions),
            observation_length=3,
            action_length=2,
        )
    )
