from __future__ import annotations

from pathlib import Path

import torch

from rootbuddy_rl.league import (
    LeaguePool,
    discover_snapshots,
    load_frozen_agent,
    parse_checkpoint_update,
)
from rootbuddy_rl.model import CandidatePolicyValueNet


def test_parse_checkpoint_update_reads_numbered_checkpoint_name() -> None:
    assert parse_checkpoint_update(Path("checkpoint_000123.pt")) == 123


def test_discover_snapshots_sorts_numbered_checkpoints(tmp_path: Path) -> None:
    save_checkpoint(tmp_path / "checkpoint_000010.pt", update=10)
    save_checkpoint(tmp_path / "checkpoint_000002.pt", update=2)
    (tmp_path / "other.pt").write_text("ignored")

    snapshots = discover_snapshots(tmp_path)

    assert [snapshot.update for snapshot in snapshots] == [2, 10]


def test_league_pool_keeps_recent_snapshots_and_samples_at_least_one_live_faction(tmp_path: Path) -> None:
    first = save_checkpoint(tmp_path / "checkpoint_000001.pt", update=1)
    second = save_checkpoint(tmp_path / "checkpoint_000002.pt", update=2)
    third = save_checkpoint(tmp_path / "checkpoint_000003.pt", update=3)
    pool = LeaguePool(max_snapshots=2, seed=707)

    pool.add_snapshot(first)
    pool.add_snapshot(second)
    pool.add_snapshot(third)
    opponents = pool.sample_faction_opponents(
        [0, 2],
        observation_dim=3,
        action_dim=2,
        opponent_fraction=1.0,
    )

    assert [snapshot.update for snapshot in pool.snapshots] == [2, 3]
    assert 0 < len(opponents) < 2
    for agent in opponents.values():
        assert not agent.model.training
        assert all(not parameter.requires_grad for parameter in agent.model.parameters())


def test_league_pool_reuses_loaded_agents(tmp_path: Path) -> None:
    checkpoint = save_checkpoint(tmp_path / "checkpoint_000001.pt", update=1)
    pool = LeaguePool(max_snapshots=2, seed=707)
    snapshot = pool.add_snapshot(checkpoint)

    first = pool.agent_for_snapshot(snapshot, observation_dim=3, action_dim=2)
    second = pool.agent_for_snapshot(snapshot, observation_dim=3, action_dim=2)

    assert first is second


def test_load_frozen_agent_uses_checkpoint_architecture(tmp_path: Path) -> None:
    checkpoint = save_checkpoint(tmp_path / "checkpoint_000004.pt", update=4)

    agent = load_frozen_agent(checkpoint, observation_dim=3, action_dim=2)

    assert agent.model.observation_dim == 3
    assert agent.model.action_dim == 2
    assert not agent.model.training
    assert all(not parameter.requires_grad for parameter in agent.model.parameters())


def save_checkpoint(path: Path, *, update: int) -> Path:
    model = CandidatePolicyValueNet(
        observation_dim=3,
        action_dim=2,
        torso_hidden_dims=(8,),
        head_hidden_dim=8,
    )
    torch.save(
        {
            "update": update,
            "config": {
                "torso_hidden_dims": (8,),
                "head_hidden_dim": 8,
                "num_factions": 4,
            },
            "model_state_dict": model.state_dict(),
        },
        path,
    )
    return path
