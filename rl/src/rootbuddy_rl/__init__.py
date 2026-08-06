"""Python bridge for RootBuddy RL training."""

from .engine_client import (
    ConfigResponse,
    EngineClient,
    EngineClientError,
    EnvBatch,
    EnvDecision,
    Faction,
    ProtocolError,
    VecEnvConfig,
)
from .model import (
    ActionEvaluation,
    ActionSelection,
    CandidatePolicyValueNet,
    ModelOutput,
    mask_logits,
    masked_categorical,
    tensors_from_batch,
)
from .vec_env import RootBuddyVecEnv, VecEnvArrays, batch_to_arrays, sample_random_actions

__all__ = [
    "ActionEvaluation",
    "ActionSelection",
    "CandidatePolicyValueNet",
    "ConfigResponse",
    "EngineClient",
    "EngineClientError",
    "EnvBatch",
    "EnvDecision",
    "Faction",
    "ModelOutput",
    "ProtocolError",
    "RootBuddyVecEnv",
    "VecEnvConfig",
    "VecEnvArrays",
    "batch_to_arrays",
    "mask_logits",
    "masked_categorical",
    "sample_random_actions",
    "tensors_from_batch",
]
