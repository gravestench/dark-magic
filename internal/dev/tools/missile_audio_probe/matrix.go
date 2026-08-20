package main

// soundSpec is the immutable sound expectation for one matrix row; loop is record metadata, not an observation.
type soundSpec struct {
	record string
	role   string
	loop   bool
}

// caseSpec locks a probe case to the skill, missile, outcome, target count, and expected sound records under test.
type caseSpec struct {
	id, skill, missile, outcome string
	skillID, targets            int
	sounds                      []soundSpec
}

// requiredCases is deliberately ordered so missing-case reports remain stable and reviewable across captures.
var requiredCases = []caseSpec{
	{
		id: "fire-bolt-expire", skillID: 36, skill: "Fire Bolt", missile: "firebolt", outcome: "expired", targets: 0,
		sounds: []soundSpec{
			{"sorceress_firebolt_1", "travel", true},
			{"sorceress_firebolt_impact_1", "hit", false},
		},
	},
	{
		id: "fire-bolt-hit", skillID: 36, skill: "Fire Bolt", missile: "firebolt", outcome: "hit", targets: 1,
		sounds: []soundSpec{
			{"sorceress_firebolt_1", "travel", true},
			{"sorceress_firebolt_impact_1", "hit", false},
		},
	},
	{
		id: "fire-ball-hit", skillID: 47, skill: "Fire Ball", missile: "fireball", outcome: "hit", targets: 1,
		sounds: []soundSpec{
			{"sorceress_fireball_1", "travel", true},
			{"sorceress_fireball_impact_1", "hit", false},
		},
	},
	{
		id: "nova-empty", skillID: 48, skill: "Nova", missile: "nova", outcome: "expired", targets: 0,
		sounds: []soundSpec{{"sorceress_nova", "travel", false}},
	},
	{
		id: "nova-three-targets", skillID: 48, skill: "Nova", missile: "nova", outcome: "multi-contact",
		targets: 3, sounds: []soundSpec{{"sorceress_nova", "travel", false}},
	},
	{
		id: "ice-blast-hit", skillID: 45, skill: "Ice Blast", missile: "iceblast", outcome: "hit", targets: 1,
		sounds: []soundSpec{
			{"sorceress_icebolt_1", "travel", true},
			{"sorceress_iceblast_impact_1", "hit", false},
		},
	},
	{
		id: "glacial-spike-hit", skillID: 55, skill: "Glacial Spike", missile: "glacialspike", outcome: "hit",
		targets: 1,
		sounds: []soundSpec{
			{"sorceress_glacialspike_1", "travel", true},
			{"sorceress_iceblast_impact_1", "hit", false},
		},
	},
}

// specFor resolves a case ID against the target-locked matrix; callers use the boolean to reject unknown rows.
func specFor(id string) (caseSpec, bool) {
	for _, spec := range requiredCases {
		if spec.id == id {
			return spec, true
		}
	}

	return caseSpec{}, false
}

// soundSpecFor resolves record metadata within one case so normalization cannot borrow expectations from another row.
func soundSpecFor(spec caseSpec, record string) (soundSpec, bool) {
	for _, sound := range spec.sounds {
		if sound.record == record {
			return sound, true
		}
	}

	return soundSpec{}, false
}
