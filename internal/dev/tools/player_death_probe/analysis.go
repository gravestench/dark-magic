package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
)

// analyze validates an exact capture before deriving a report, so invalid evidence never receives a fingerprint.
func analyze(input io.Reader) (report, error) {
	captured, data, err := decodeCapture(input)
	if err != nil {
		return report{}, err
	}

	if err := validate(captured); err != nil {
		return report{}, err
	}

	fingerprint := sha256.Sum256(data)

	result := report{
		Schema:             probeSchema + ".report",
		Target:             probeTarget,
		ExecutableSHA256:   captured.Runtime.ExecutableSHA256,
		CaptureFingerprint: hex.EncodeToString(fingerprint[:]),
	}
	for _, observed := range captured.Cases {
		result.Cases = append(result.Cases, normalize(observed))
	}

	return result, nil
}

// decodeCapture retains the original bytes for provenance while accepting exactly one strict JSON value.
func decodeCapture(input io.Reader) (capture, []byte, error) {
	data, err := io.ReadAll(input)
	if err != nil {
		return capture{}, nil, fmt.Errorf("player death probe: read capture: %w", err)
	}

	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()

	var captured capture
	if err := decoder.Decode(&captured); err != nil {
		return capture{}, nil, fmt.Errorf("player death probe: decode capture: %w", err)
	}

	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return capture{}, nil, fmt.Errorf("player death probe: capture must contain one JSON value")
	}

	return captured, data, nil
}
