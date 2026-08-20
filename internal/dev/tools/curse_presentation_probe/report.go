package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
)

// report is the normalized output contract. Field declaration order intentionally controls human-readable JSON order.
type report struct {
	Schema             string       `json:"schema"`
	Target             string       `json:"target"`
	ExecutableSHA256   string       `json:"executable_sha256"`
	CaptureFingerprint string       `json:"capture_fingerprint"`
	Coverage           coverage     `json:"coverage"`
	Cases              []caseReport `json:"cases"`
}

// coverage reports whether both supported skills were observed across every target-count band.
type coverage struct {
	Complete bool     `json:"complete"`
	Missing  []string `json:"missing,omitempty"`
}

// caseReport preserves capture order while replacing absolute instance coordinates with anchor-relative offsets.
type caseReport struct {
	ID         string        `json:"id"`
	SkillID    int           `json:"skill_id"`
	TargetBand string        `json:"target_band"`
	Layers     []layerReport `json:"layers"`
}

// layerReport preserves the observed layer and instance ordering in normalized output.
type layerReport struct {
	MissileRecord string           `json:"missile_record"`
	Present       bool             `json:"present"`
	Instances     []instanceReport `json:"instances"`
}

// instanceReport records duration, anchor-relative motion, and the original tracking declaration.
type instanceReport struct {
	FirstFrame   int    `json:"first_frame"`
	LastFrame    int    `json:"last_frame"`
	Frames       int    `json:"frames"`
	Anchor       string `json:"anchor"`
	TargetIndex  *int   `json:"target_index,omitempty"`
	StartOffset  point  `json:"start_offset"`
	EndOffset    point  `json:"end_offset"`
	Translated   bool   `json:"translated"`
	TracksAnchor bool   `json:"tracks_anchor"`
}

// buildReport fingerprints the original capture and normalizes cases without reordering observations. The resulting
// report therefore remains traceable to exact input bytes while retaining capture chronology.
func buildReport(captured capture, rawCapture []byte) report {
	fingerprint := sha256.Sum256(rawCapture)
	result := report{
		Schema:             probeSchema + ".report",
		Target:             probeTarget,
		ExecutableSHA256:   captured.Runtime.ExecutableSHA256,
		CaptureFingerprint: hex.EncodeToString(fingerprint[:]),
		Coverage:           coverageFor(captured.Cases),
	}

	for _, observed := range captured.Cases {
		result.Cases = append(result.Cases, normalize(observed))
	}

	return result
}

// normalize converts one validated case to anchor-relative coordinates while preserving every layer and instance in
// observed order.
func normalize(observed probeCase) caseReport {
	result := caseReport{
		ID:         observed.ID,
		SkillID:    observed.SkillID,
		TargetBand: targetBand(len(observed.Targets)),
	}

	for _, current := range observed.Layers {
		layerResult := layerReport{
			MissileRecord: current.MissileRecord,
			Present:       current.Present,
		}
		for _, item := range current.Instances {
			anchor := anchorPoint(observed, item)
			layerResult.Instances = append(layerResult.Instances, instanceReport{
				FirstFrame:   item.FirstFrame,
				LastFrame:    item.LastFrame,
				Frames:       item.LastFrame - item.FirstFrame + 1,
				Anchor:       item.Anchor,
				TargetIndex:  item.TargetIndex,
				StartOffset:  subtract(item.Start, anchor),
				EndOffset:    subtract(item.End, anchor),
				Translated:   item.Start != item.End,
				TracksAnchor: item.TracksAnchor,
			})
		}

		result.Layers = append(result.Layers, layerResult)
	}

	return result
}

// coverageFor computes the fixed skill-by-target-band matrix and sorts gaps so reports stay deterministic regardless
// of capture order.
func coverageFor(cases []probeCase) coverage {
	seen := make(map[string]bool)
	for _, observed := range cases {
		seen[fmt.Sprintf("skill-%d:%s", observed.SkillID, targetBand(len(observed.Targets)))] = true
	}

	result := coverage{}

	for _, skillID := range []int{66, 72} {
		for _, band := range []string{"empty", "single", "multiple"} {
			key := fmt.Sprintf("skill-%d:%s", skillID, band)
			if !seen[key] {
				result.Missing = append(result.Missing, key)
			}
		}
	}

	sort.Strings(result.Missing)
	result.Complete = len(result.Missing) == 0

	return result
}

// expectedMissiles maps supported skill IDs to the exact record names required from owned-runtime evidence.
func expectedMissiles(skillID int) ([]string, string, bool) {
	switch skillID {
	case 66:
		return []string{"curseamplifydamage", "cursecast"}, "Amplify Damage", true
	case 72:
		return []string{"curseweaken", "cursecast"}, "Weaken", true
	default:
		return nil, "", false
	}
}

// targetBand reduces exact target counts to the coverage categories promised by the report schema.
func targetBand(count int) string {
	if count == 0 {
		return "empty"
	}

	if count == 1 {
		return "single"
	}

	return "multiple"
}

// anchorPoint resolves a validated instance anchor. Panicking on an unknown anchor protects normalization from being
// used without the validation phase that guarantees target indexes and anchor names.
func anchorPoint(observed probeCase, item instance) point {
	switch item.Anchor {
	case "caster":
		return observed.Caster
	case "cursor":
		return observed.Cursor
	case "target":
		return observed.Targets[*item.TargetIndex]
	default:
		panic("validated anchor")
	}
}

// subtract expresses an observed point relative to its resolved anchor, which makes motion comparable across captures.
func subtract(value, origin point) point {
	return point{X: value.X - origin.X, Y: value.Y - origin.Y}
}
