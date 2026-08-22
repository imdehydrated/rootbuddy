"""Checkpoint-backed opponent pool for self-play stabilization."""

from __future__ import annotations

import re
import time
from dataclasses import dataclass
from pathlib import Path
from typing import Iterable

import numpy as np
import torch

from .model import CandidatePolicyValueNet, tensors_from_batch
from .train_constants import FACTION_COUNT
from .vec_env import VecEnvArrays

CHECKPOINT_PATTERN = re.compile(r"checkpoint_(\d+)\.pt$")


@dataclass(frozen=True)
class LeagueSnapshot:
    path: Path
    update: int
    created_at: float


@dataclass
class FrozenModelAgent:
    model: CandidatePolicyValueNet
    deterministic: bool = True
    device: str | torch.device | None = None

    def __post_init__(self) -> None:
        self.model.eval()
        for parameter in self.model.parameters():
            parameter.requires_grad_(False)

    def select_actions(self, batch: VecEnvArrays) -> np.ndarray:
        if np.any(batch.candidate_counts <= 0):
            raise ValueError("frozen model cannot select from rows without legal actions")
        tensors = tensors_from_batch(batch, device=self.device)
        with torch.no_grad():
            selection = self.model.act(
                tensors["observations"],
                tensors["candidate_actions"],
                tensors["action_mask"],
                tensors["active_factions"],
                deterministic=self.deterministic,
            )
        return selection.actions.detach().cpu().numpy().astype(np.int64)


class LeaguePool:
    def __init__(
        self,
        *,
        max_snapshots: int = 8,
        seed: int = 0,
        deterministic_agents: bool = True,
    ) -> None:
        if max_snapshots <= 0:
            raise ValueError("max_snapshots must be positive")
        self.max_snapshots = max_snapshots
        self.deterministic_agents = deterministic_agents
        self.rng = np.random.default_rng(seed if seed != 0 else None)
        self.snapshots: list[LeagueSnapshot] = []
        self._agent_cache: dict[tuple[Path, int, int, str | None], FrozenModelAgent] = {}

    def add_snapshot(self, path: str | Path, *, update: int | None = None) -> LeagueSnapshot:
        snapshot_path = Path(path)
        snapshot = LeagueSnapshot(
            path=snapshot_path,
            update=parse_checkpoint_update(snapshot_path) if update is None else int(update),
            created_at=time.time(),
        )
        self.snapshots = [existing for existing in self.snapshots if existing.path != snapshot.path]
        self.snapshots.append(snapshot)
        self.snapshots.sort(key=lambda item: item.update)
        if len(self.snapshots) > self.max_snapshots:
            removed = self.snapshots[: len(self.snapshots) - self.max_snapshots]
            self.snapshots = self.snapshots[-self.max_snapshots :]
            for old in removed:
                self._drop_cached_agents(old.path)
        return snapshot

    def extend(self, snapshots: Iterable[LeagueSnapshot]) -> None:
        for snapshot in snapshots:
            self.add_snapshot(snapshot.path, update=snapshot.update)

    def sample_faction_opponents(
        self,
        factions: Iterable[int],
        *,
        observation_dim: int,
        action_dim: int,
        opponent_fraction: float,
        device: str | torch.device | None = None,
    ) -> dict[int, FrozenModelAgent]:
        if opponent_fraction <= 0 or len(self.snapshots) == 0:
            return {}
        faction_ids = [int(faction) for faction in factions]
        if len(faction_ids) <= 1:
            return {}

        selected = [faction for faction in faction_ids if self.rng.random() < opponent_fraction]
        if len(selected) >= len(faction_ids):
            selected.pop(int(self.rng.integers(0, len(selected))))
        if len(selected) == 0 and opponent_fraction >= 1.0:
            selected = [faction_ids[int(self.rng.integers(0, len(faction_ids)))]]

        opponents: dict[int, FrozenModelAgent] = {}
        for faction in selected:
            snapshot = self.sample_snapshot()
            opponents[faction] = self.agent_for_snapshot(
                snapshot,
                observation_dim=observation_dim,
                action_dim=action_dim,
                device=device,
            )
        return opponents

    def sample_snapshot(self) -> LeagueSnapshot:
        if len(self.snapshots) == 0:
            raise ValueError("cannot sample from empty league pool")
        index = int(self.rng.integers(0, len(self.snapshots)))
        return self.snapshots[index]

    def agent_for_snapshot(
        self,
        snapshot: LeagueSnapshot,
        *,
        observation_dim: int,
        action_dim: int,
        device: str | torch.device | None = None,
    ) -> FrozenModelAgent:
        key = (snapshot.path, observation_dim, action_dim, str(device) if device is not None else None)
        cached = self._agent_cache.get(key)
        if cached is not None:
            return cached
        agent = load_frozen_agent(
            snapshot.path,
            observation_dim=observation_dim,
            action_dim=action_dim,
            deterministic=self.deterministic_agents,
            device=device,
        )
        self._agent_cache[key] = agent
        return agent

    def _drop_cached_agents(self, path: Path) -> None:
        for key in list(self._agent_cache):
            if key[0] == path:
                del self._agent_cache[key]


def load_frozen_agent(
    checkpoint_path: str | Path,
    *,
    observation_dim: int,
    action_dim: int,
    deterministic: bool = True,
    device: str | torch.device | None = None,
) -> FrozenModelAgent:
    checkpoint = torch.load(checkpoint_path, map_location=device, weights_only=False)
    config = dict(checkpoint.get("config") or {})
    model = CandidatePolicyValueNet(
        observation_dim=observation_dim,
        action_dim=action_dim,
        torso_hidden_dims=tuple(config.get("torso_hidden_dims", (256, 256))),
        head_hidden_dim=int(config.get("head_hidden_dim", 128)),
        num_factions=int(config.get("num_factions", FACTION_COUNT)),
    )
    model.load_state_dict(checkpoint["model_state_dict"])
    if device is not None:
        model.to(device)
    return FrozenModelAgent(model=model, deterministic=deterministic, device=device)


def discover_snapshots(directory: str | Path) -> tuple[LeagueSnapshot, ...]:
    root = Path(directory)
    if not root.exists():
        return ()
    snapshots = []
    for path in root.glob("checkpoint_*.pt"):
        try:
            update = parse_checkpoint_update(path)
        except ValueError:
            continue
        snapshots.append(LeagueSnapshot(path=path, update=update, created_at=path.stat().st_mtime))
    snapshots.sort(key=lambda item: item.update)
    return tuple(snapshots)


def parse_checkpoint_update(path: str | Path) -> int:
    match = CHECKPOINT_PATTERN.search(Path(path).name)
    if match is None:
        raise ValueError(f"checkpoint path does not contain an update number: {path}")
    return int(match.group(1))
