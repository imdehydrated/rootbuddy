from __future__ import annotations

import argparse
from pathlib import Path

import torch

from rootbuddy_rl.engine_client import VecEnvConfig
from rootbuddy_rl.evaluate import EvaluationMetrics
from rootbuddy_rl.experiment import (
    apply_overrides,
    diagnose_experiment,
    experiment_preset,
    format_experiment_result,
    run_experiment,
    win_rate_gap,
)
from rootbuddy_rl.model import CandidatePolicyValueNet
from rootbuddy_rl.train import TrainConfig, TrainMetrics, TrainResult


def test_experiment_preset_builds_smoke_train_eval_cycle(tmp_path: Path) -> None:
    config = experiment_preset("smoke", output_dir=tmp_path)

    assert config.name == "smoke"
    assert config.train_config.updates == 2
    assert config.train_config.checkpoint_every == 1
    assert config.train_config.league_opponent_fraction > 0
    assert config.train_config.env_config.track_all_hands
    assert config.train_config.env_config.step_penalty == 0.01
    assert config.train_config.env_config.truncation_penalty == 10.0
    assert config.eval_episodes == 2
    assert config.baselines == ("random", "greedy-vp")
    assert Path(config.train_config.checkpoint_dir).is_relative_to(tmp_path)


def test_two_player_presets_use_eval_sized_step_caps(tmp_path: Path) -> None:
    fast = experiment_preset("two-player-fast", output_dir=tmp_path)
    stable = experiment_preset("two-player-stable", output_dir=tmp_path)

    assert fast.train_config.env_config.max_steps == 1024
    assert stable.train_config.env_config.max_steps == 2048


def test_apply_overrides_changes_tuning_knobs(tmp_path: Path) -> None:
    config = experiment_preset("smoke", output_dir=tmp_path)
    args = argparse.Namespace(
        updates=7,
        rollout_steps=9,
        learning_rate=1e-4,
        entropy_coef=0.03,
        terminal_win_bonus=20.0,
        step_penalty=0.02,
        truncation_penalty=12.0,
        league_opponent_fraction=0.5,
        device="cpu",
        output_dir=str(tmp_path / "runs"),
        eval_episodes=6,
        eval_num_envs=3,
        baselines="random",
    )

    updated = apply_overrides(config, args)

    assert updated.train_config.updates == 7
    assert updated.train_config.rollout_steps == 9
    assert updated.train_config.learning_rate == 1e-4
    assert updated.train_config.entropy_coef == 0.03
    assert updated.train_config.env_config.terminal_win_bonus == 20.0
    assert updated.train_config.env_config.step_penalty == 0.02
    assert updated.train_config.env_config.truncation_penalty == 12.0
    assert updated.train_config.league_opponent_fraction == 0.5
    assert updated.train_config.device == "cpu"
    assert updated.eval_episodes == 6
    assert updated.eval_num_envs == 3
    assert updated.baselines == ("random",)
    assert Path(updated.train_config.checkpoint_dir).is_relative_to(tmp_path / "runs")


def test_diagnose_experiment_flags_unstable_training_and_eval() -> None:
    diagnostics = diagnose_experiment(
        [
            train_metric(entropy=0.5, approx_kl=0.3),
            train_metric(entropy=0.01, approx_kl=0.1),
        ],
        {
            "random": EvaluationMetrics(
                episodes=4,
                transitions=16,
                truncated_episodes=2,
                mean_reward=0.0,
                mean_game_length=4.0,
                transitions_per_second=10.0,
                per_faction_win_rate=(1.0, 0.0, 0.0, 0.0),
                per_faction_terminal_vp=(30.0, 4.0, 2.0, 0.0),
            )
        },
        max_truncation_rate=0.25,
        min_entropy=0.05,
        max_approx_kl=0.2,
        max_win_rate_gap=0.5,
    )

    assert diagnostics.final_entropy == 0.01
    assert diagnostics.max_approx_kl == 0.3
    assert diagnostics.max_truncation_rate == 0.5
    assert diagnostics.max_win_rate_gap == 1.0
    assert len(diagnostics.warnings) == 4


def test_run_experiment_trains_then_evaluates_each_baseline() -> None:
    calls: list[str] = []
    config = experiment_preset("smoke")

    def fake_train(train_config: TrainConfig) -> TrainResult:
        calls.append(f"train:{train_config.updates}")
        model = CandidatePolicyValueNet(
            observation_dim=3,
            action_dim=2,
            torso_hidden_dims=(8,),
            head_hidden_dim=8,
        )
        optimizer = torch.optim.Adam(model.parameters(), lr=1e-3)
        return TrainResult(model=model, optimizer=optimizer, metrics=(train_metric(),))

    def fake_evaluate(eval_config, *, agents, default_agent, env=None):
        calls.append(f"eval:{eval_config.episodes}:{type(default_agent).__name__}:{sorted(agents)}")
        return EvaluationMetrics(
            episodes=eval_config.episodes,
            transitions=4,
            truncated_episodes=0,
            mean_reward=0.25,
            mean_game_length=2.0,
            transitions_per_second=100.0,
            per_faction_win_rate=(0.5, 0.0, 0.5, 0.0),
            per_faction_terminal_vp=(10.0, 0.0, 12.0, 0.0),
        )

    result = run_experiment(config, train_fn=fake_train, evaluate_fn=fake_evaluate)

    assert calls == [
        "train:2",
        "eval:2:RandomAgent:[0]",
        "eval:2:GreedyVPAgent:[0]",
    ]
    assert set(result.evaluations) == {"random", "greedy-vp"}
    assert result.diagnostics.warnings == ()
    assert "eval[random]" in format_experiment_result(result)


def test_win_rate_gap_uses_observed_faction_rates() -> None:
    assert win_rate_gap((0.5, 0.25, 0.25, 0.0)) == 0.5


def train_metric(*, entropy: float = 0.5, approx_kl: float = 0.01) -> TrainMetrics:
    return TrainMetrics(
        update=1,
        rollout_steps=4,
        transitions=8,
        completed_episodes=1,
        mean_reward=0.2,
        mean_game_length=4.0,
        per_faction_win_rate=(0.5, 0.0, 0.5, 0.0),
        per_faction_terminal_vp=(8.0, 0.0, 9.0, 0.0),
        transitions_per_second=100.0,
        loss=0.1,
        policy_loss=0.01,
        value_loss=0.2,
        entropy=entropy,
        approx_kl=approx_kl,
        clip_fraction=0.0,
        checkpoint_path=None,
    )
