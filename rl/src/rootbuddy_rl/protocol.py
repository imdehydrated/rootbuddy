"""Length-prefixed JSON framing shared by the RootBuddy RL client."""

from __future__ import annotations

import json
import struct
from typing import Any, BinaryIO

MAX_FRAME_BYTES = 256 * 1024 * 1024
_HEADER = struct.Struct(">I")


class ProtocolError(RuntimeError):
    """Raised when the stdio bridge emits or receives invalid frames."""


def read_frame(stream: BinaryIO, *, max_frame_bytes: int = MAX_FRAME_BYTES) -> bytes:
    header = stream.read(_HEADER.size)
    if header == b"":
        raise EOFError("stream closed before frame header")
    if len(header) != _HEADER.size:
        raise ProtocolError(f"incomplete frame header: got {len(header)} bytes")

    (length,) = _HEADER.unpack(header)
    if length > max_frame_bytes:
        raise ProtocolError(f"frame length {length} exceeds max {max_frame_bytes}")

    payload = stream.read(length)
    if len(payload) != length:
        raise ProtocolError(f"incomplete frame payload: got {len(payload)} of {length} bytes")
    return payload


def write_frame(stream: BinaryIO, payload: bytes) -> None:
    if len(payload) > 0xFFFFFFFF:
        raise ProtocolError(f"payload length {len(payload)} exceeds uint32 frame limit")
    stream.write(_HEADER.pack(len(payload)))
    stream.write(payload)
    stream.flush()


def read_json_frame(stream: BinaryIO) -> dict[str, Any]:
    payload = read_frame(stream)
    try:
        decoded = json.loads(payload.decode("utf-8"))
    except (UnicodeDecodeError, json.JSONDecodeError) as exc:
        raise ProtocolError(f"invalid JSON frame: {exc}") from exc
    if not isinstance(decoded, dict):
        raise ProtocolError(f"expected JSON object frame, got {type(decoded).__name__}")
    return decoded


def write_json_frame(stream: BinaryIO, message: dict[str, Any]) -> None:
    payload = json.dumps(message, separators=(",", ":")).encode("utf-8")
    write_frame(stream, payload)
