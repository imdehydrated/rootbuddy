"""Client for the Go `cmd/rlenv` length-prefixed stdio server."""

from __future__ import annotations

import os
import subprocess
from dataclasses import dataclass
from enum import IntEnum
from pathlib import Path
from typing import Any, Iterable, Sequence

import numpy as np

from .protocol import ProtocolError, read_json_frame, write_json_frame


class EngineClientError(RuntimeError):
    """Raised when the Go environment rejects a request."""


class Faction(IntEnum):
    MARQUISE = 0
    ALLIANCE = 1
    EYRIE = 2
    VAGABOND = 3


@dataclass(frozen=True)
class VecEnvConfig:
    num_envs: int
    base_seed: int
    max_steps: int = 0
    factions: Sequence[int | Faction] | None = None
    player_faction: int | Faction = Faction.MARQUISE
    map_id: str = "autumn"
    track_all_hands: bool = True
    terminal_win_bonus: float = 0.0
    step_penalty: float = 0.0
    truncation_penalty: float = 0.0

    def to_wire(self) -> dict[str, Any]:
        wire = {
            "numEnvs": self.num_envs,
            "baseSeed": self.base_seed,
            "maxSteps": self.max_steps,
            "playerFaction": int(self.player_faction),
            "mapId": self.map_id,
            "trackAllHands": self.track_all_hands,
            "terminalWinBonus": self.terminal_win_bonus,
            "stepPenalty": self.step_penalty,
            "truncationPenalty": self.truncation_penalty,
        }
        if self.factions is not None:
            wire["factions"] = [int(faction) for faction in self.factions]
        return wire


@dataclass(frozen=True)
class EnvDecision:
    env_index: int
    episode: int
    step: int
    active_faction: int
    current_phase: int
    current_step: int
    observation: np.ndarray
    candidate_actions: np.ndarray
    candidate_rewards: np.ndarray
    candidate_count: int
    reward: float
    done: bool
    truncated: bool
    winner: int
    winning_coalition: tuple[int, ...]
    victory_points: tuple[int, ...]
    legal_action_types: tuple[int, ...]
    observation_length: int
    action_length: int


@dataclass(frozen=True)
class EnvBatch:
    decisions: tuple[EnvDecision, ...]
    observation_length: int
    action_length: int

    @property
    def rewards(self) -> np.ndarray:
        return np.asarray([decision.reward for decision in self.decisions], dtype=np.float32)

    @property
    def dones(self) -> np.ndarray:
        return np.asarray([decision.done for decision in self.decisions], dtype=np.bool_)

    @property
    def active_factions(self) -> np.ndarray:
        return np.asarray([decision.active_faction for decision in self.decisions], dtype=np.int64)

    @property
    def candidate_counts(self) -> np.ndarray:
        return np.asarray([decision.candidate_count for decision in self.decisions], dtype=np.int64)


@dataclass(frozen=True)
class ConfigResponse:
    config: dict[str, Any]
    observation_length: int
    action_length: int


class EngineClient:
    def __init__(
        self,
        command: Sequence[str] | None = None,
        *,
        cwd: str | Path | None = None,
    ) -> None:
        self.command = list(command) if command is not None else ["go", "run", "./cmd/rlenv"]
        self.cwd = Path(cwd) if cwd is not None else _repo_root()
        self._process: subprocess.Popen[bytes] | None = None

    def __enter__(self) -> EngineClient:
        self.start()
        return self

    def __exit__(self, exc_type: object, exc: object, traceback: object) -> None:
        self.close()

    def start(self) -> None:
        if self._process is not None:
            return
        self._process = subprocess.Popen(
            self.command,
            cwd=self.cwd,
            env=_process_env(self.cwd),
            stdin=subprocess.PIPE,
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
        )

    def close(self) -> None:
        process = self._process
        if process is None:
            return
        self._process = None
        if process.stdin is not None:
            process.stdin.close()
        if process.poll() is None:
            process.terminate()
            try:
                process.wait(timeout=5)
            except subprocess.TimeoutExpired:
                process.kill()
                process.wait(timeout=5)

    def configure(self, config: VecEnvConfig | dict[str, Any]) -> ConfigResponse:
        wire_config = config.to_wire() if isinstance(config, VecEnvConfig) else dict(config)
        response = self._request({"type": "config", "config": wire_config})
        return ConfigResponse(
            config=dict(response.get("config") or {}),
            observation_length=int(response.get("observationLength", 0)),
            action_length=int(response.get("actionLength", 0)),
        )

    def reset(self, env_indices: Iterable[int] | None = None) -> EnvBatch:
        request: dict[str, Any] = {"type": "reset"}
        if env_indices is not None:
            request["envIndices"] = list(env_indices)
        return self._parse_batch(self._request(request))

    def step(self, action_indices: Iterable[int]) -> EnvBatch:
        response = self._request({"type": "step", "actionIndices": list(action_indices)})
        return self._parse_batch(response)

    def _request(self, message: dict[str, Any]) -> dict[str, Any]:
        self.start()
        process = self._require_process()
        assert process.stdin is not None
        assert process.stdout is not None

        write_json_frame(process.stdin, message)
        try:
            response = read_json_frame(process.stdout)
        except EOFError as exc:
            stderr = _read_available_stderr(process)
            detail = f"; stderr: {stderr}" if stderr else ""
            raise ProtocolError(f"rlenv closed stdout before response{detail}") from exc

        if not response.get("ok", False):
            raise EngineClientError(str(response.get("error", "unknown rlenv error")))
        return response

    def _parse_batch(self, response: dict[str, Any]) -> EnvBatch:
        observation_length = int(response.get("observationLength", 0))
        action_length = int(response.get("actionLength", 0))
        decisions = tuple(
            _parse_decision(raw, observation_length=observation_length, action_length=action_length)
            for raw in response.get("decisions", [])
        )
        return EnvBatch(
            decisions=decisions,
            observation_length=observation_length,
            action_length=action_length,
        )

    def _require_process(self) -> subprocess.Popen[bytes]:
        if self._process is None:
            raise RuntimeError("EngineClient process has not been started")
        if self._process.poll() is not None:
            raise ProtocolError(f"rlenv exited with code {self._process.returncode}")
        return self._process


def _parse_decision(raw: dict[str, Any], *, observation_length: int, action_length: int) -> EnvDecision:
    observation = np.asarray(raw.get("observation", []), dtype=np.float32)
    candidates = np.asarray(raw.get("candidateActions", []), dtype=np.float32)
    rewards = np.asarray(raw.get("candidateRewards", []), dtype=np.float32)
    if candidates.size == 0:
        candidates = candidates.reshape((0, action_length))
    return EnvDecision(
        env_index=int(raw.get("envIndex", 0)),
        episode=int(raw.get("episode", 0)),
        step=int(raw.get("step", 0)),
        active_faction=int(raw.get("activeFaction", 0)),
        current_phase=int(raw.get("currentPhase", 0)),
        current_step=int(raw.get("currentStep", 0)),
        observation=observation,
        candidate_actions=candidates,
        candidate_rewards=rewards,
        candidate_count=int(raw.get("candidateCount", 0)),
        reward=float(raw.get("reward", 0.0)),
        done=bool(raw.get("done", False)),
        truncated=bool(raw.get("truncated", False)),
        winner=int(raw.get("winner", 0)),
        winning_coalition=tuple(int(faction) for faction in raw.get("winningCoalition", [])),
        victory_points=tuple(int(points) for points in raw.get("victoryPoints", [])),
        legal_action_types=tuple(int(action_type) for action_type in raw.get("legalActionTypes", [])),
        observation_length=int(raw.get("observationLength", observation_length)),
        action_length=int(raw.get("actionLength", action_length)),
    )


def _repo_root() -> Path:
    return Path(__file__).resolve().parents[3]


def _process_env(cwd: Path) -> dict[str, str]:
    env = os.environ.copy()
    if "GOCACHE" not in env:
        cache_dir = cwd / ".gocache"
        cache_dir.mkdir(parents=True, exist_ok=True)
        env["GOCACHE"] = str(cache_dir)
    return env


def _read_available_stderr(process: subprocess.Popen[bytes]) -> str:
    if process.stderr is None:
        return ""
    if process.poll() is None:
        try:
            process.wait(timeout=1)
        except subprocess.TimeoutExpired:
            return ""
    try:
        return process.stderr.read().decode("utf-8", errors="replace").strip()
    except OSError:
        return ""
