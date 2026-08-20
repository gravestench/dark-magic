package main

import (
	"fmt"
	"sort"
)

type report struct {
	Schema             string       `json:"schema"`
	Target             string       `json:"target"`
	ExecutableSHA256   string       `json:"executable_sha256"`
	RuntimeSession     string       `json:"runtime_session"`
	Records            records      `json:"records"`
	CaptureFingerprint string       `json:"capture_fingerprint"`
	Cases              []caseReport `json:"cases"`
}

type caseReport struct {
	ID              string          `json:"id"`
	Trigger         string          `json:"trigger"`
	Difficulty      string          `json:"difficulty"`
	SkillLevel      int             `json:"skill_level"`
	PlayerLevel     int             `json:"player_level"`
	MapSeed         uint32          `json:"map_seed"`
	ExplosionCenter point           `json:"explosion_center"`
	OrderedEvents   []orderedEvent  `json:"ordered_events"`
	Targets         []targetReport  `json:"targets"`
	AffectedKinds   []string        `json:"affected_kinds"`
	UnaffectedKinds []string        `json:"unaffected_kinds"`
	RadiusBrackets  []radiusBracket `json:"radius_brackets"`
}

type orderedEvent struct {
	Name  string `json:"name"`
	Frame int    `json:"frame"`
}

type targetReport struct {
	ID                 string   `json:"id"`
	Kind               string   `json:"kind"`
	HostileToOwner     bool     `json:"hostile_to_owner"`
	Position           point    `json:"position"`
	DistanceMilli      int      `json:"distance_millisubtiles"`
	HealthDeltaRaw     int64    `json:"health_delta_raw"`
	FireResistance     int      `json:"fire_resistance_percent"`
	PhysicalResistance int      `json:"physical_resistance_percent"`
	Channels           channels `json:"pre_mitigation_channels_raw"`
	DamageEvent        bool     `json:"damage_event"`
	HitReaction        bool     `json:"hit_reaction"`
	Died               bool     `json:"died"`
}

type radiusBracket struct {
	Kind                   string `json:"kind"`
	HostileToOwner         bool   `json:"hostile_to_owner"`
	FarthestAffectedMilli  int    `json:"farthest_affected_millisubtiles"`
	NearestUnaffectedMilli int    `json:"nearest_unaffected_millisubtiles"`
	BoundaryBracketed      bool   `json:"boundary_bracketed"`
}

type targetProfileRange struct {
	kind              string
	hostile           bool
	farthestAffected  int
	nearestUnaffected int
}

type targetSummary struct {
	reports         []targetReport
	affectedKinds   map[string]bool
	unaffectedKinds map[string]bool
	radiusProfiles  map[string]*targetProfileRange
}

// normalize turns one validated observation into deterministic evidence without reordering the target samples.
func normalize(observed probeCase) caseReport {
	result := caseReport{
		ID:              observed.ID,
		Trigger:         observed.Trigger,
		Difficulty:      observed.Difficulty,
		SkillLevel:      observed.SkillLevel,
		PlayerLevel:     observed.PlayerLevel,
		MapSeed:         observed.MapSeed,
		ExplosionCenter: observed.ExplosionCenter,
		OrderedEvents:   normalizedEventOrder(observed.EventFrames),
	}

	summary := summarizeTargets(observed.Targets)
	result.Targets = summary.reports
	result.AffectedKinds = sortedKeys(summary.affectedKinds)
	result.UnaffectedKinds = sortedKeys(summary.unaffectedKinds)
	result.RadiusBrackets = normalizedRadiusBrackets(summary.radiusProfiles)

	return result
}

// normalizedEventOrder sorts events by frame while stable ties preserve their removal, explosion, creation meaning.
func normalizedEventOrder(frames eventFrames) []orderedEvent {
	events := []orderedEvent{
		{Name: "old_golem_removed", Frame: *frames.OldGolemRemoved},
		{Name: "explosion_started", Frame: *frames.ExplosionStarted},
	}
	if frames.NewGolemCreated != nil {
		events = append(events, orderedEvent{Name: "new_golem_created", Frame: *frames.NewGolemCreated})
	}

	sort.SliceStable(
		events,
		// The strict comparison makes equal-frame events retain the semantic order encoded above.
		func(left, right int) bool {
			return events[left].Frame < events[right].Frame
		},
	)

	return events
}

// summarizeTargets preserves sample order while collecting deterministic kind and radius evidence for presentation.
func summarizeTargets(samples []targetSample) targetSummary {
	summary := targetSummary{
		affectedKinds:   make(map[string]bool),
		unaffectedKinds: make(map[string]bool),
		radiusProfiles:  make(map[string]*targetProfileRange),
	}

	for _, sample := range samples {
		summary.reports = append(summary.reports, normalizedTarget(sample))
		if !sample.DamageEvent {
			summary.unaffectedKinds[sample.Kind] = true
			continue
		}

		summary.affectedKinds[sample.Kind] = true
		profileKey := targetProfileKey(sample)

		profile := summary.radiusProfiles[profileKey]
		if profile == nil {
			profile = &targetProfileRange{
				kind:              sample.Kind,
				hostile:           sample.HostileToOwner,
				farthestAffected:  -1,
				nearestUnaffected: -1,
			}
			summary.radiusProfiles[profileKey] = profile
		}

		if sample.DistanceMilli > profile.farthestAffected {
			profile.farthestAffected = sample.DistanceMilli
		}
	}

	bracketNearestUnaffected(samples, summary.radiusProfiles)

	return summary
}

// normalizedTarget derives the observed health delta while copying every remaining measurement unchanged.
func normalizedTarget(sample targetSample) targetReport {
	return targetReport{
		ID:                 sample.ID,
		Kind:               sample.Kind,
		HostileToOwner:     sample.HostileToOwner,
		Position:           sample.Position,
		DistanceMilli:      sample.DistanceMilli,
		HealthDeltaRaw:     sample.HealthBeforeRaw - sample.HealthAfterRaw,
		FireResistance:     sample.FireResistance,
		PhysicalResistance: sample.PhysicalResistance,
		Channels:           sample.Channels,
		DamageEvent:        sample.DamageEvent,
		HitReaction:        sample.HitReaction,
		Died:               sample.Died,
	}
}

// bracketNearestUnaffected bounds each affected profile only with a sample from the same kind and hostility class.
func bracketNearestUnaffected(samples []targetSample, profiles map[string]*targetProfileRange) {
	for _, sample := range samples {
		if sample.DamageEvent {
			continue
		}

		// A nearby owner or ally proves the target filter, not the radius, when its profile differs.
		profile := profiles[targetProfileKey(sample)]
		if profile == nil {
			continue
		}

		if profile.nearestUnaffected < 0 || sample.DistanceMilli < profile.nearestUnaffected {
			profile.nearestUnaffected = sample.DistanceMilli
		}
	}
}

// normalizedRadiusBrackets sorts profile keys so map iteration cannot make serialized reports nondeterministic.
func normalizedRadiusBrackets(profiles map[string]*targetProfileRange) []radiusBracket {
	profileKeys := make([]string, 0, len(profiles))
	for key := range profiles {
		profileKeys = append(profileKeys, key)
	}

	sort.Strings(profileKeys)

	var brackets []radiusBracket

	for _, key := range profileKeys {
		profile := profiles[key]
		brackets = append(brackets, radiusBracket{
			Kind:                   profile.kind,
			HostileToOwner:         profile.hostile,
			FarthestAffectedMilli:  profile.farthestAffected,
			NearestUnaffectedMilli: profile.nearestUnaffected,
			BoundaryBracketed:      profile.nearestUnaffected > profile.farthestAffected,
		})
	}

	return brackets
}

// targetProfileKey keeps hostility in the identity because matching unit kinds can belong to opposite target classes.
func targetProfileKey(sample targetSample) string {
	return fmt.Sprintf("%s:%t", sample.Kind, sample.HostileToOwner)
}

// sortedKeys converts classification sets into stable JSON arrays, including non-nil empty arrays when no keys exist.
func sortedKeys(values map[string]bool) []string {
	result := make([]string, 0, len(values))
	for value := range values {
		result = append(result, value)
	}

	sort.Strings(result)

	return result
}
