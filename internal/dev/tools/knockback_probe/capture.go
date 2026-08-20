package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
)

const (
	probeSchema = "d2legacy.knockback_probe/v1"
	probeTarget = "diablo-ii-lod-1.14d-expansion"
)

// capture is the complete owned-runtime observation envelope. Its JSON tags
// are the input contract and must remain stable for previously recorded probes.
type capture struct {
	Schema string      `json:"schema"`
	Target string      `json:"target"`
	Source string      `json:"source"`
	Cases  []probeCase `json:"cases"`
}

// probeCase groups one mechanism observation with the control context needed
// to decide whether their outcomes are comparable.
type probeCase struct {
	ID                    string  `json:"id"`
	ControlID             string  `json:"control_id,omitempty"`
	Mechanism             string  `json:"mechanism"`
	Difficulty            string  `json:"difficulty"`
	AttackerKind          string  `json:"attacker_kind"`
	Target                target  `json:"target"`
	MissileKnockbackValue int     `json:"missile_knockback_value,omitempty"`
	OpenDistanceSubtiles  float64 `json:"open_distance_subtiles"`
	Trials                []trial `json:"trials"`
}

// target records the identity and animation capability that determine whether
// a knockback chance can be inferred from the observed reactions.
type target struct {
	Kind          string `json:"kind"`
	Record        string `json:"record"`
	SizeClass     string `json:"size_class"`
	ModeSupported bool   `json:"mode_supported"`
}

// trial preserves both combat eligibility and visible motion because a valid
// knockback reaction may be blocked from producing displacement.
type trial struct {
	Hit                  bool    `json:"hit"`
	CombatBlocked        bool    `json:"combat_blocked"`
	Lethal               bool    `json:"lethal"`
	Uninterruptible      bool    `json:"uninterruptible"`
	DisplacementSubtiles float64 `json:"displacement_subtiles"`
	Reaction             string  `json:"reaction"`
}

// decodeCapture accepts exactly one strict JSON value so trailing data and
// unknown fields cannot silently corrupt the evidence being normalized.
func decodeCapture(input io.Reader) (capture, error) {
	decoder := json.NewDecoder(input)
	decoder.DisallowUnknownFields()

	var captured capture
	if err := decoder.Decode(&captured); err != nil {
		return capture{}, fmt.Errorf("knockback probe: decode capture: %w", err)
	}

	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return capture{}, fmt.Errorf("knockback probe: capture must contain one JSON value")
	}

	return captured, nil
}
