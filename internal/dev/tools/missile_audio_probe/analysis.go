package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sort"
)

// analyze decodes one capture, validates its provenance, and returns a stable report.
// Hashing the original bytes preserves an auditable link to the exact evidence rather than a re-encoded equivalent.
func analyze(input io.Reader) (report, error) {
	data, err := io.ReadAll(input)
	if err != nil {
		return report{}, fmt.Errorf("missile audio probe: read capture: %w", err)
	}

	captured, err := decodeCapture(data)
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
		RecordGenerationID: captured.Runtime.RecordGenerationID,
		CaptureFingerprint: hex.EncodeToString(fingerprint[:]),
		Coverage:           coverageFor(captured.Cases),
	}
	for _, observed := range captured.Cases {
		result.Cases = append(result.Cases, normalize(observed))
	}

	// Stable case order makes reports reviewable and prevents capture order from producing noisy diffs.
	sort.Slice(result.Cases, func(i, j int) bool {
		return result.Cases[i].ID < result.Cases[j].ID
	})

	return result, nil
}

// decodeCapture accepts exactly one strict JSON document so ignored fields cannot masquerade as evidence.
func decodeCapture(data []byte) (capture, error) {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()

	var captured capture
	if err := decoder.Decode(&captured); err != nil {
		return capture{}, fmt.Errorf("missile audio probe: decode capture: %w", err)
	}

	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return capture{}, fmt.Errorf("missile audio probe: capture must contain one JSON value")
	}

	return captured, nil
}

// normalize translates absolute capture frames into event-relative offsets that can be compared across runs.
func normalize(observed probeCase) caseReport {
	spec, _ := specFor(observed.ID)

	result := caseReport{
		ID:                    observed.ID,
		SkillID:               observed.SkillID,
		SkillLevel:            observed.SkillLevel,
		MissileRecord:         observed.MissileRecord,
		Outcome:               observed.Outcome,
		TargetCount:           observed.TargetCount,
		ProjectileVisualCount: observed.ProjectileVisualCount,
		LifetimeFrames:        observed.MissileRemovedFrame - observed.CastEffectFrame,
	}
	if observed.ContactFrame != nil {
		contactFromEffect := *observed.ContactFrame - observed.CastEffectFrame
		result.ContactFromEffect = &contactFromEffect
	}

	for _, sound := range observed.Sounds {
		result.Sounds = append(result.Sounds, normalizeSound(observed, spec, sound))
	}

	sort.Slice(result.Sounds, func(i, j int) bool {
		return result.Sounds[i].Record < result.Sounds[j].Record
	})

	return result
}

// normalizeSound retains record metadata while expressing each interval relative to visible lifecycle events.
func normalizeSound(observed probeCase, spec caseSpec, sound soundObservation) soundReport {
	expected, _ := soundSpecFor(spec, sound.Record)
	normalized := soundReport{
		Record:     sound.Record,
		Role:       sound.Role,
		RecordLoop: expected.loop,
		Present:    sound.Present,
		Instances:  len(sound.Intervals),
	}

	for _, interval := range sound.Intervals {
		item := intervalReport{
			FirstFromEffect: interval.FirstFrame - observed.CastEffectFrame,
			LastFromEffect:  interval.LastFrame - observed.CastEffectFrame,
			LastFromRemoval: interval.LastFrame - observed.MissileRemovedFrame,
		}
		if observed.ContactFrame != nil {
			firstFromContact := interval.FirstFrame - *observed.ContactFrame
			item.FirstFromContact = &firstFromContact
		}

		normalized.Intervals = append(normalized.Intervals, item)
	}

	return normalized
}

// coverageFor identifies missing target-locked cases without treating a partial capture as complete evidence.
func coverageFor(cases []probeCase) coverage {
	seen := make(map[string]bool, len(cases))
	for _, observed := range cases {
		seen[observed.ID] = true
	}

	result := coverage{Complete: true}

	for _, required := range requiredCases {
		if !seen[required.id] {
			result.Complete = false
			result.Missing = append(result.Missing, required.id)
		}
	}

	return result
}
