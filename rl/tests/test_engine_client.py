from __future__ import annotations

import numpy as np

from rootbuddy_rl.engine_client import VecEnvConfig, _parse_decision


def test_vec_env_config_uses_go_json_keys() -> None:
    config = VecEnvConfig(
        num_envs=2,
        base_seed=707,
        max_steps=10,
        factions=[0, 2],
        player_faction=0,
        track_all_hands=True,
    )

    assert config.to_wire() == {
        "numEnvs": 2,
        "baseSeed": 707,
        "maxSteps": 10,
        "playerFaction": 0,
        "mapId": "autumn",
        "trackAllHands": True,
        "terminalWinBonus": 0.0,
        "stepPenalty": 0.0,
        "truncationPenalty": 0.0,
        "factions": [0, 2],
    }


def test_vec_env_config_tracks_all_hands_by_default() -> None:
    config = VecEnvConfig(num_envs=1, base_seed=707)

    assert config.to_wire()["trackAllHands"] is True


def test_parse_decision_converts_vectors_to_numpy() -> None:
    decision = _parse_decision(
        {
            "envIndex": 1,
            "episode": 0,
            "step": 3,
            "activeFaction": 2,
            "currentPhase": 1,
            "currentStep": 3,
            "observation": [0.0, 1.0],
            "candidateActions": [[1.0, 0.0], [0.0, 1.0]],
            "candidateRewards": [0.0, 1.0],
            "candidateCount": 2,
            "reward": 1.5,
            "done": False,
            "truncated": False,
            "winner": 0,
            "winningCoalition": [],
            "victoryPoints": [1, 2, 3, 4],
            "legalActionTypes": [1, 4],
            "observationLength": 2,
            "actionLength": 2,
        },
        observation_length=2,
        action_length=2,
    )

    assert decision.env_index == 1
    assert decision.current_phase == 1
    assert decision.current_step == 3
    assert decision.candidate_count == 2
    assert decision.observation.dtype == np.float32
    assert decision.candidate_actions.shape == (2, 2)
    assert decision.candidate_rewards.tolist() == [0.0, 1.0]
    assert not decision.truncated
    assert decision.victory_points == (1, 2, 3, 4)
