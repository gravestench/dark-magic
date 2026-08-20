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

// analyze validates one byte-exact capture before deriving an ordered report whose fingerprint identifies that input.
func analyze(input io.Reader) (report, error) {
	data, err := io.ReadAll(input)
	if err != nil {
		return report{}, fmt.Errorf("defense outcome probe: read capture: %w", err)
	}

	captured, err := decodeCapture(data)
	if err != nil {
		return report{}, err
	}

	if err := validateCapture(captured); err != nil {
		return report{}, err
	}

	// Hash the original bytes rather than a re-encoding so whitespace or field-order changes remain detectable.
	fingerprint := sha256.Sum256(data)
	result := report{
		Schema:             probeSchema + ".report",
		Target:             probeTarget,
		ExecutableSHA256:   captured.Runtime.ExecutableSHA256,
		CaptureFingerprint: hex.EncodeToString(fingerprint[:]),
	}

	controlsByID := make(map[string]probeCase, len(captured.Cases))
	for _, observed := range captured.Cases {
		controlsByID[observed.ID] = observed
	}

	// Preserve capture order in the report; consumers use that order to compare each mechanism with its control.
	for _, observed := range captured.Cases {
		result.Cases = append(result.Cases, normalizeCase(observed, controlsByID[observed.ControlID]))
	}

	return result, nil
}

// decodeCapture accepts exactly one strict JSON value so ignored fields or appended data cannot contaminate evidence.
func decodeCapture(data []byte) (capture, error) {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()

	var captured capture
	if err := decoder.Decode(&captured); err != nil {
		return capture{}, fmt.Errorf("defense outcome probe: decode capture: %w", err)
	}

	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return capture{}, fmt.Errorf("defense outcome probe: capture must contain one JSON value")
	}

	return captured, nil
}
