package movement

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
)

// MoveCommand identifies the replayable command understood by the d2legacy movement simulation.
const MoveCommand = "player.move"

// MovePayload carries either discrete directional input or an exact world-space target.
// A target takes precedence when the simulation resolves the command.
type MovePayload struct {
	X       int         `json:"x"`
	Y       int         `json:"y"`
	Running bool        `json:"running"`
	Target  *MoveTarget `json:"target,omitempty"`
}

// decodeMove validates a serialized movement command without accepting schema extensions.
// Rejecting unknown and trailing data keeps replay behavior identical across runtime versions.
func decodeMove(encoded []byte) (MovePayload, error) {
	var payload MovePayload

	decoder := json.NewDecoder(bytes.NewReader(encoded))
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&payload); err != nil {
		return MovePayload{}, err
	}

	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return MovePayload{}, fmt.Errorf("movement payload has trailing data")
	}

	if payload.X < -1 || payload.X > 1 || payload.Y < -1 || payload.Y > 1 {
		return MovePayload{}, fmt.Errorf("movement axes must be between -1 and 1")
	}

	if !validMoveTarget(payload.Target) {
		return MovePayload{}, fmt.Errorf("movement target must be finite")
	}

	return payload, nil
}

// validMoveTarget rejects non-finite coordinates before they can make simulation math non-deterministic.
func validMoveTarget(target *MoveTarget) bool {
	if target == nil {
		return true
	}

	return !math.IsNaN(target.X) && !math.IsNaN(target.Y) &&
		!math.IsInf(target.X, 0) && !math.IsInf(target.Y, 0)
}
