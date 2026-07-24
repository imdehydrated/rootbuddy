package main

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"io"
	"strings"
	"testing"

	"github.com/imdehydrated/rootbuddy/engine/rl"
	"github.com/imdehydrated/rootbuddy/game"
)

func TestProtocolConfigResetStep(t *testing.T) {
	input := &bytes.Buffer{}
	appendRequestFrame(t, input, requestEnvelope{
		Type: "config",
		Config: &rl.VecEnvConfig{
			NumEnvs:       2,
			BaseSeed:      707,
			MaxSteps:      10,
			Factions:      []game.Faction{game.Marquise, game.Eyrie},
			PlayerFaction: game.Marquise,
		},
	})
	appendRequestFrame(t, input, requestEnvelope{Type: "reset"})
	appendRequestFrame(t, input, requestEnvelope{
		Type:          "step",
		ActionIndices: []int{0, 0},
	})

	output := &bytes.Buffer{}
	if err := run(input, output); err != nil {
		t.Fatalf("run failed: %v", err)
	}

	config := readResponseFrame(t, output)
	if !config.OK {
		t.Fatalf("config response failed: %s", config.Error)
	}
	if config.Config == nil || config.Config.NumEnvs != 2 {
		t.Fatalf("config response = %+v, want normalized config with two envs", config.Config)
	}
	if config.ObservationLength != rl.ObservationVectorLength() || config.ActionLength != rl.ActionVectorLength() {
		t.Fatalf("config lengths = %d/%d, want %d/%d", config.ObservationLength, config.ActionLength, rl.ObservationVectorLength(), rl.ActionVectorLength())
	}

	reset := readResponseFrame(t, output)
	if !reset.OK {
		t.Fatalf("reset response failed: %s", reset.Error)
	}
	if len(reset.Decisions) != 2 {
		t.Fatalf("reset decisions = %d, want 2", len(reset.Decisions))
	}
	for _, decision := range reset.Decisions {
		if decision.CandidateCount == 0 {
			t.Fatalf("reset decision has no candidates: %+v", decision)
		}
		if len(decision.Observation) != rl.ObservationVectorLength() {
			t.Fatalf("reset observation length = %d, want %d", len(decision.Observation), rl.ObservationVectorLength())
		}
	}

	step := readResponseFrame(t, output)
	if !step.OK {
		t.Fatalf("step response failed: %s", step.Error)
	}
	if len(step.Decisions) != 2 {
		t.Fatalf("step decisions = %d, want 2", len(step.Decisions))
	}
	for _, decision := range step.Decisions {
		if decision.Step != 1 {
			t.Fatalf("step decision step = %d, want 1", decision.Step)
		}
	}
	assertNoMoreFrames(t, output)
}

func TestProtocolRejectsResetBeforeConfig(t *testing.T) {
	input := &bytes.Buffer{}
	appendRequestFrame(t, input, requestEnvelope{Type: "reset"})

	output := &bytes.Buffer{}
	if err := run(input, output); err != nil {
		t.Fatalf("run failed: %v", err)
	}

	response := readResponseFrame(t, output)
	if response.OK || !strings.Contains(response.Error, "config first") {
		t.Fatalf("response = %+v, want config-first error", response)
	}
}

func TestProtocolRejectsInvalidStepCount(t *testing.T) {
	input := &bytes.Buffer{}
	appendRequestFrame(t, input, requestEnvelope{
		Type: "config",
		Config: &rl.VecEnvConfig{
			NumEnvs:       2,
			BaseSeed:      808,
			Factions:      []game.Faction{game.Marquise, game.Eyrie},
			PlayerFaction: game.Marquise,
		},
	})
	appendRequestFrame(t, input, requestEnvelope{Type: "reset"})
	appendRequestFrame(t, input, requestEnvelope{
		Type:          "step",
		ActionIndices: []int{0},
	})

	output := &bytes.Buffer{}
	if err := run(input, output); err != nil {
		t.Fatalf("run failed: %v", err)
	}

	_ = readResponseFrame(t, output)
	_ = readResponseFrame(t, output)
	response := readResponseFrame(t, output)
	if response.OK || !strings.Contains(response.Error, "action index count") {
		t.Fatalf("response = %+v, want action count error", response)
	}
}

func TestProtocolRejectsUnknownMessageType(t *testing.T) {
	input := &bytes.Buffer{}
	appendRequestFrame(t, input, requestEnvelope{Type: "noop"})

	output := &bytes.Buffer{}
	if err := run(input, output); err != nil {
		t.Fatalf("run failed: %v", err)
	}

	response := readResponseFrame(t, output)
	if response.OK || !strings.Contains(response.Error, "unknown message type") {
		t.Fatalf("response = %+v, want unknown message error", response)
	}
}

func appendRequestFrame(t *testing.T, buffer *bytes.Buffer, request requestEnvelope) {
	t.Helper()

	payload, err := json.Marshal(request)
	if err != nil {
		t.Fatalf("marshal request failed: %v", err)
	}
	var header [4]byte
	binary.BigEndian.PutUint32(header[:], uint32(len(payload)))
	buffer.Write(header[:])
	buffer.Write(payload)
}

func readResponseFrame(t *testing.T, reader io.Reader) responseEnvelope {
	t.Helper()

	payload, err := readFrame(reader)
	if err != nil {
		t.Fatalf("read response failed: %v", err)
	}
	var response responseEnvelope
	if err := json.Unmarshal(payload, &response); err != nil {
		t.Fatalf("unmarshal response failed: %v", err)
	}
	return response
}

func assertNoMoreFrames(t *testing.T, reader io.Reader) {
	t.Helper()

	if _, err := readFrame(reader); err != io.EOF {
		t.Fatalf("readFrame error = %v, want EOF", err)
	}
}
