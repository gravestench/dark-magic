package main

import (
	"bytes"
	"cmp"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"slices"
)

// analyze reads and validates one capture before deriving a report, so invalid evidence never reaches normalization.
func analyze(input io.Reader) (report, error) {
	data, err := io.ReadAll(input)
	if err != nil {
		return report{}, fmt.Errorf("cast-rate probe: read capture: %w", err)
	}

	captured, err := decodeCapture(data)
	if err != nil {
		return report{}, err
	}

	if err := validate(captured); err != nil {
		return report{}, err
	}

	return reportFor(captured, data), nil
}

// decodeCapture accepts exactly one strict JSON value, preventing ignored fields or trailing data from tainting proof.
func decodeCapture(data []byte) (capture, error) {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()

	var captured capture
	if err := decoder.Decode(&captured); err != nil {
		return capture{}, fmt.Errorf("cast-rate probe: decode capture: %w", err)
	}

	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return capture{}, fmt.Errorf("cast-rate probe: capture must contain one JSON value")
	}

	return captured, nil
}

// reportFor converts validated evidence into deterministic profiles while fingerprinting the original capture bytes.
func reportFor(captured capture, encodedCapture []byte) report {
	fingerprint := sha256.Sum256(encodedCapture)
	result := report{
		Schema:             probeSchema + ".report",
		Target:             probeTarget,
		ExecutableSHA256:   captured.Runtime.ExecutableSHA256,
		RecordGenerationID: captured.Runtime.RecordGenerationID,
		CaptureFingerprint: hex.EncodeToString(fingerprint[:]),
		Coverage:           coverageFor(captured.Cases),
	}

	for _, observed := range captured.Cases {
		result.Profiles = append(result.Profiles, reportProfileFor(observed))
	}

	// Stable field precedence makes reports reproducible even when captures list cases in a different order.
	slices.SortFunc(result.Profiles, compareReportProfiles)

	return result
}

// reportProfileFor derives delays from absolute frames without retaining capture-only provenance fields.
func reportProfileFor(observed probeCase) caseReport {
	return caseReport{
		ID:                observed.ID,
		SkillID:           observed.SkillID,
		AnimationMode:     observed.AnimationMode,
		SequenceNumber:    observed.SequenceNumber,
		WeaponClass:       observed.WeaponClass,
		RawFasterCastRate: observed.RawFasterCastRate,
		EffectDelay:       observed.EffectFrame - observed.StartFrame,
		CompletionDelay:   observed.NeutralFrame - observed.StartFrame,
	}
}

// compareReportProfiles preserves the report's established skill, weapon, rate, and ID ordering.
func compareReportProfiles(left, right caseReport) int {
	if order := cmp.Compare(left.SkillID, right.SkillID); order != 0 {
		return order
	}

	if order := cmp.Compare(left.WeaponClass, right.WeaponClass); order != 0 {
		return order
	}

	if order := cmp.Compare(left.RawFasterCastRate, right.RawFasterCastRate); order != 0 {
		return order
	}

	return cmp.Compare(left.ID, right.ID)
}
