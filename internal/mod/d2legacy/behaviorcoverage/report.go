// Package behaviorcoverage inventories target Skills.txt server behavior
// signatures without inferring implementation support from resemblance.
package behaviorcoverage

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"
)

const (
	Schema       = "d2legacy.skill_behavior_coverage/v1"
	ReportSchema = "d2legacy.skill_behavior_coverage.report/v1"
	Target       = "diablo-ii-lod-1.14d-expansion"
)

type Manifest struct {
	Schema          string           `json:"schema"`
	Version         int              `json:"version"`
	Target          string           `json:"target"`
	Implementations []Implementation `json:"implementations"`
}

type Implementation struct {
	SkillID        int    `json:"skill_id"`
	Family         string `json:"family"`
	EvidenceStatus string `json:"evidence_status"`
}

type Report struct {
	Schema    string          `json:"schema"`
	Target    string          `json:"target"`
	Sources   Sources         `json:"sources"`
	Summary   Summary         `json:"summary"`
	Behaviors []BehaviorGroup `json:"behaviors"`
}

type Sources struct {
	SkillsTable   string `json:"skills_table"`
	SkillsLayer   string `json:"skills_layer"`
	MissilesTable string `json:"missiles_table"`
	MissilesLayer string `json:"missiles_layer"`
}

type Summary struct {
	SkillRows         int `json:"skill_rows"`
	BehaviorGroups    int `json:"behavior_groups"`
	ImplementedSkills int `json:"implemented_skills"`
	MissingSkills     int `json:"missing_skills"`
}

type BehaviorGroup struct {
	ServerStartFunction      string     `json:"server_start_function"`
	ServerDoFunction         string     `json:"server_do_function"`
	MissileServerDoFunctions []string   `json:"missile_server_do_functions"`
	Consumers                []Consumer `json:"consumers"`
}

type Consumer struct {
	SkillID              int                `json:"skill_id"`
	Skill                string             `json:"skill"`
	ServerMissiles       []MissileReference `json:"server_missiles"`
	ImplementationFamily string             `json:"implementation_family,omitempty"`
	MissingFamily        bool               `json:"missing_family"`
	EvidenceStatus       string             `json:"evidence_status"`
}

type MissileReference struct {
	Slot             string `json:"slot"`
	Missile          string `json:"missile"`
	ServerDoFunction string `json:"server_do_function"`
}

// DecodeManifest accepts exactly one strict JSON value and validates its target.
// Rejecting trailing or unknown data keeps coverage declarations reviewable.
func DecodeManifest(input io.Reader) (Manifest, error) {
	decoder := json.NewDecoder(input)
	decoder.DisallowUnknownFields()

	var manifest Manifest
	if err := decoder.Decode(&manifest); err != nil {
		return Manifest{}, fmt.Errorf("skill behavior coverage: decode manifest: %w", err)
	}

	if err := decoder.Decode(new(any)); !errors.Is(err, io.EOF) {
		return Manifest{}, fmt.Errorf("skill behavior coverage: manifest must contain one JSON value")
	}

	if err := manifest.Validate(); err != nil {
		return Manifest{}, err
	}

	return manifest, nil
}

// Validate enforces the manifest's target and one-to-one skill-family mapping.
// Builds can therefore treat every declaration as trusted implementation evidence.
func (m Manifest) Validate() error {
	if m.Schema != Schema {
		return fmt.Errorf("skill behavior coverage: schema %q, want %q", m.Schema, Schema)
	}

	if m.Version != 1 {
		return fmt.Errorf("skill behavior coverage: version %d, want 1", m.Version)
	}

	if m.Target != Target {
		return fmt.Errorf("skill behavior coverage: target %q, want %q", m.Target, Target)
	}

	seen := make(map[int]bool, len(m.Implementations))
	for _, implementation := range m.Implementations {
		if implementation.SkillID < 0 || seen[implementation.SkillID] {
			return fmt.Errorf("skill behavior coverage: invalid or duplicate skill ID %d", implementation.SkillID)
		}

		seen[implementation.SkillID] = true
		if strings.TrimSpace(implementation.Family) == "" || strings.TrimSpace(implementation.EvidenceStatus) == "" {
			return fmt.Errorf("skill behavior coverage: skill %d requires family and evidence status", implementation.SkillID)
		}
	}

	return nil
}

// Build groups target skill rows by their complete server behavior signature.
// Only explicit manifest declarations count as implemented; similar table rows
// deliberately remain missing so the report never invents coverage.
func Build(manifest Manifest, skills, missiles []map[string]string, sources Sources) (Report, error) {
	if err := manifest.Validate(); err != nil {
		return Report{}, err
	}

	implemented := indexImplementations(manifest.Implementations)
	missileRows := indexMissiles(missiles)
	groups := make(map[string]*BehaviorGroup)
	foundImplementations := make(map[int]bool, len(implemented))

	for rowNumber, skill := range skills {
		consumer, functions, err := buildConsumer(skill, rowNumber, missileRows)
		if err != nil {
			return Report{}, err
		}

		if implementation, ok := implemented[consumer.SkillID]; ok {
			consumer.ImplementationFamily = implementation.Family
			consumer.MissingFamily = false
			consumer.EvidenceStatus = implementation.EvidenceStatus
			foundImplementations[consumer.SkillID] = true
		}

		addConsumerToGroup(groups, skill, functions, consumer)
	}

	for skillID := range implemented {
		if !foundImplementations[skillID] {
			return Report{}, fmt.Errorf(
				"skill behavior coverage: declared skill %d is absent from mounted Skills.txt",
				skillID,
			)
		}
	}

	return assembleReport(skills, groups, foundImplementations, sources), nil
}

// indexImplementations creates the explicit coverage lookup. Validation has
// already guaranteed unique IDs, so later entries cannot silently overwrite.
func indexImplementations(implementations []Implementation) map[int]Implementation {
	result := make(map[int]Implementation, len(implementations))
	for _, implementation := range implementations {
		result[implementation.SkillID] = implementation
	}

	return result
}

// indexMissiles normalizes table names for Diablo II's case-insensitive joins.
// Empty names remain absent, matching the previous table interpretation.
func indexMissiles(missiles []map[string]string) map[string]map[string]string {
	result := make(map[string]map[string]string, len(missiles))
	for _, row := range missiles {
		name := strings.TrimSpace(row["Missile"])
		if name != "" {
			result[strings.ToLower(name)] = row
		}
	}

	return result
}

// buildConsumer converts one Skills.txt row and its server missiles into the
// full evidence consumer. Missing missile rows fail instead of weakening a key.
func buildConsumer(
	skill map[string]string,
	rowNumber int,
	missileRows map[string]map[string]string,
) (Consumer, []string, error) {
	id, err := strconv.Atoi(strings.TrimSpace(skill["Id"]))
	if err != nil {
		return Consumer{}, nil, fmt.Errorf(
			"skill behavior coverage: Skills.txt row %d has invalid Id %q",
			rowNumber+2,
			skill["Id"],
		)
	}

	consumer := Consumer{
		SkillID:        id,
		Skill:          strings.TrimSpace(skill["skill"]),
		ServerMissiles: make([]MissileReference, 0, 4),
		MissingFamily:  true,
		EvidenceStatus: "missing-implementation-family",
	}
	missileFunctions := map[string]bool{}

	for _, field := range serverMissileFields() {
		name := strings.TrimSpace(skill[field.column])
		if name == "" {
			continue
		}

		missile, found := missileRows[strings.ToLower(name)]
		if !found {
			return Consumer{}, nil, fmt.Errorf(
				"skill behavior coverage: skill %d references missing server missile %q",
				id,
				name,
			)
		}

		function := strings.TrimSpace(missile["pSrvDoFunc"])
		consumer.ServerMissiles = append(consumer.ServerMissiles, MissileReference{
			Slot: field.slot, Missile: name, ServerDoFunction: function,
		})
		missileFunctions[function] = true
	}

	return consumer, sortedKeys(missileFunctions), nil
}

type missileField struct {
	column string
	slot   string
}

// serverMissileFields preserves the canonical primary/a/b/c slot order used in
// report JSON. That order is evidence and must not follow map iteration.
func serverMissileFields() []missileField {
	return []missileField{
		{column: "srvmissile", slot: "primary"},
		{column: "srvmissilea", slot: "a"},
		{column: "srvmissileb", slot: "b"},
		{column: "srvmissilec", slot: "c"},
	}
}

// addConsumerToGroup constructs a collision-safe signature key and appends the
// consumer without reordering source rows inside the group.
func addConsumerToGroup(
	groups map[string]*BehaviorGroup,
	skill map[string]string,
	missileFunctions []string,
	consumer Consumer,
) {
	start := strings.TrimSpace(skill["srvstfunc"])
	do := strings.TrimSpace(skill["srvdofunc"])
	key := start + "\x00" + do + "\x00" + strings.Join(missileFunctions, "\x00")

	group := groups[key]
	if group == nil {
		group = &BehaviorGroup{
			ServerStartFunction:      start,
			ServerDoFunction:         do,
			MissileServerDoFunctions: missileFunctions,
			Consumers:                make([]Consumer, 0),
		}
		groups[key] = group
	}

	group.Consumers = append(group.Consumers, consumer)
}

// assembleReport sorts behavior keys and consumers so output remains identical
// across map iteration orders, then derives summary counts from emitted data.
func assembleReport(
	skills []map[string]string,
	groups map[string]*BehaviorGroup,
	foundImplementations map[int]bool,
	sources Sources,
) Report {
	result := Report{Schema: ReportSchema, Target: Target, Sources: sources}

	keys := sortedKeys(groups)
	for _, key := range keys {
		group := groups[key]
		sort.Slice(group.Consumers, func(i, j int) bool {
			return group.Consumers[i].SkillID < group.Consumers[j].SkillID
		})
		result.Behaviors = append(result.Behaviors, *group)
	}

	result.Summary = Summary{
		SkillRows:         len(skills),
		BehaviorGroups:    len(result.Behaviors),
		ImplementedSkills: len(foundImplementations),
		MissingSkills:     len(skills) - len(foundImplementations),
	}

	return result
}

// sortedKeys returns deterministic map-key order for signatures and missile
// function sets, preventing report churn between otherwise identical runs.
func sortedKeys[T any](values map[string]T) []string {
	result := make([]string, 0, len(values))
	for value := range values {
		result = append(result, value)
	}

	sort.Strings(result)

	return result
}
