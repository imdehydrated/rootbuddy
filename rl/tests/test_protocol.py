from __future__ import annotations

import io
import json
import struct

import pytest

from rootbuddy_rl.protocol import ProtocolError, read_json_frame, write_json_frame


def test_json_frame_round_trip() -> None:
    stream = io.BytesIO()
    write_json_frame(stream, {"type": "reset", "value": 7})

    stream.seek(0)
    assert read_json_frame(stream) == {"type": "reset", "value": 7}


def test_read_json_frame_rejects_non_object() -> None:
    payload = json.dumps(["not", "object"]).encode("utf-8")
    stream = io.BytesIO(struct.pack(">I", len(payload)) + payload)

    with pytest.raises(ProtocolError, match="expected JSON object"):
        read_json_frame(stream)
