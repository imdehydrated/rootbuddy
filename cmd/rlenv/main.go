package main

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/imdehydrated/rootbuddy/engine/rl"
)

const maxFrameBytes = 256 * 1024 * 1024
const maxUint32 = uint64(^uint32(0))

type requestEnvelope struct {
	Type          string           `json:"type"`
	Config        *rl.VecEnvConfig `json:"config,omitempty"`
	ActionIndices []int            `json:"actionIndices,omitempty"`
	EnvIndices    []int            `json:"envIndices,omitempty"`
}

type responseEnvelope struct {
	Type              string           `json:"type"`
	OK                bool             `json:"ok"`
	Error             string           `json:"error,omitempty"`
	Config            *rl.VecEnvConfig `json:"config,omitempty"`
	Decisions         []rl.EnvDecision `json:"decisions,omitempty"`
	ObservationLength int              `json:"observationLength,omitempty"`
	ActionLength      int              `json:"actionLength,omitempty"`
}

type protocolServer struct {
	env *rl.VecEnv
}

func main() {
	if err := run(os.Stdin, os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "rlenv: %v\n", err)
		os.Exit(1)
	}
}

func run(input io.Reader, output io.Writer) error {
	server := &protocolServer{}
	for {
		payload, err := readFrame(input)
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return err
		}

		response := server.handlePayload(payload)
		if err := writeFrame(output, response); err != nil {
			return err
		}
	}
}

func (server *protocolServer) handlePayload(payload []byte) responseEnvelope {
	var request requestEnvelope
	if err := json.Unmarshal(payload, &request); err != nil {
		return errorResponse("", fmt.Errorf("decode request: %w", err))
	}
	return server.handleRequest(request)
}

func (server *protocolServer) handleRequest(request requestEnvelope) responseEnvelope {
	switch request.Type {
	case "config":
		return server.handleConfig(request)
	case "reset":
		return server.handleReset(request)
	case "step":
		return server.handleStep(request)
	default:
		return errorResponse(request.Type, fmt.Errorf("unknown message type %q", request.Type))
	}
}

func (server *protocolServer) handleConfig(request requestEnvelope) responseEnvelope {
	if request.Config == nil {
		return errorResponse(request.Type, errors.New("config message requires config"))
	}

	env, err := rl.NewVecEnv(*request.Config)
	if err != nil {
		return errorResponse(request.Type, err)
	}
	server.env = env
	config := env.Config()
	return responseEnvelope{
		Type:              request.Type,
		OK:                true,
		Config:            &config,
		ObservationLength: rl.ObservationVectorLength(),
		ActionLength:      rl.ActionVectorLength(),
	}
}

func (server *protocolServer) handleReset(request requestEnvelope) responseEnvelope {
	if server.env == nil {
		return errorResponse("reset", errors.New("reset requires config first"))
	}

	var decisions []rl.EnvDecision
	if len(request.EnvIndices) == 0 {
		result, err := server.env.Reset()
		if err != nil {
			return errorResponse("reset", err)
		}
		decisions = result.Decisions
	} else {
		decisions = make([]rl.EnvDecision, len(request.EnvIndices))
		for row, index := range request.EnvIndices {
			decision, err := server.env.ResetEnv(index)
			if err != nil {
				return errorResponse("reset", err)
			}
			decisions[row] = decision
		}
	}
	return responseEnvelope{
		Type:              "reset",
		OK:                true,
		Decisions:         decisions,
		ObservationLength: rl.ObservationVectorLength(),
		ActionLength:      rl.ActionVectorLength(),
	}
}

func (server *protocolServer) handleStep(request requestEnvelope) responseEnvelope {
	if server.env == nil {
		return errorResponse(request.Type, errors.New("step requires config first"))
	}

	result, err := server.env.Step(rl.StepRequest{ActionIndices: request.ActionIndices})
	if err != nil {
		return errorResponse(request.Type, err)
	}
	return responseEnvelope{
		Type:              request.Type,
		OK:                true,
		Decisions:         result.Decisions,
		ObservationLength: rl.ObservationVectorLength(),
		ActionLength:      rl.ActionVectorLength(),
	}
}

func errorResponse(messageType string, err error) responseEnvelope {
	return responseEnvelope{
		Type:  messageType,
		OK:    false,
		Error: err.Error(),
	}
}

func readFrame(reader io.Reader) ([]byte, error) {
	var header [4]byte
	if _, err := io.ReadFull(reader, header[:]); err != nil {
		if errors.Is(err, io.EOF) {
			return nil, io.EOF
		}
		return nil, err
	}

	length := binary.BigEndian.Uint32(header[:])
	if length > maxFrameBytes {
		return nil, fmt.Errorf("frame length %d exceeds max %d", length, maxFrameBytes)
	}

	payload := make([]byte, length)
	if _, err := io.ReadFull(reader, payload); err != nil {
		return nil, err
	}
	return payload, nil
}

func writeFrame(writer io.Writer, response responseEnvelope) error {
	payload, err := json.Marshal(response)
	if err != nil {
		return fmt.Errorf("encode response: %w", err)
	}
	if uint64(len(payload)) > maxUint32 {
		return fmt.Errorf("response length %d exceeds uint32 frame limit", len(payload))
	}

	var header [4]byte
	binary.BigEndian.PutUint32(header[:], uint32(len(payload)))
	if err := writeAll(writer, header[:]); err != nil {
		return err
	}
	return writeAll(writer, payload)
}

func writeAll(writer io.Writer, payload []byte) error {
	for len(payload) > 0 {
		written, err := writer.Write(payload)
		if err != nil {
			return err
		}
		if written == 0 {
			return io.ErrShortWrite
		}
		payload = payload[written:]
	}
	return nil
}
