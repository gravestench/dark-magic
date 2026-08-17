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

func Build(manifest Manifest, skills, missiles []map[string]string, sources Sources) (Report, error) {
	if err := manifest.Validate(); err != nil {
		return Report{}, err
	}
	implemented := make(map[int]Implementation, len(manifest.Implementations))
	for _, implementation := range manifest.Implementations {
		implemented[implementation.SkillID] = implementation
	}
	missileRows := make(map[string]map[string]string, len(missiles))
	for _, row := range missiles {
		name := strings.TrimSpace(row["Missile"])
		if name != "" {
			missileRows[strings.ToLower(name)] = row
		}
	}

	groups := make(map[string]*BehaviorGroup)
	foundImplementations := make(map[int]bool, len(implemented))
	for rowNumber, skill := range skills {
		id, err := strconv.Atoi(strings.TrimSpace(skill["Id"]))
		if err != nil {
			return Report{}, fmt.Errorf("skill behavior coverage: Skills.txt row %d has invalid Id %q", rowNumber+2, skill["Id"])
		}
		consumer := Consumer{
			SkillID:        id,
			Skill:          strings.TrimSpace(skill["skill"]),
			ServerMissiles: make([]MissileReference, 0, 4),
			MissingFamily:  true,
			EvidenceStatus: "missing-implementation-family",
		}
		missileFunctions := map[string]bool{}
		for _, field := range []struct{ column, slot string }{
			{"srvmissile", "primary"},
			{"srvmissilea", "a"},
			{"srvmissileb", "b"},
			{"srvmissilec", "c"},
		} {
			name := strings.TrimSpace(skill[field.column])
			if name == "" {
				continue
			}
			missile, ok := missileRows[strings.ToLower(name)]
			if !ok {
				return Report{}, fmt.Errorf("skill behavior coverage: skill %d references missing server missile %q", id, name)
			}
			function := strings.TrimSpace(missile["pSrvDoFunc"])
			consumer.ServerMissiles = append(consumer.ServerMissiles, MissileReference{
				Slot: field.slot, Missile: name, ServerDoFunction: function,
			})
			missileFunctions[function] = true
		}
		functions := sortedKeys(missileFunctions)
		if implementation, ok := implemented[id]; ok {
			consumer.ImplementationFamily = implementation.Family
			consumer.MissingFamily = false
			consumer.EvidenceStatus = implementation.EvidenceStatus
			foundImplementations[id] = true
		}
		start, do := strings.TrimSpace(skill["srvstfunc"]), strings.TrimSpace(skill["srvdofunc"])
		key := start + "\x00" + do + "\x00" + strings.Join(functions, "\x00")
		group := groups[key]
		if group == nil {
			group = &BehaviorGroup{
				ServerStartFunction: start, ServerDoFunction: do,
				MissileServerDoFunctions: functions,
				Consumers:                make([]Consumer, 0),
			}
			groups[key] = group
		}
		group.Consumers = append(group.Consumers, consumer)
	}
	for skillID := range implemented {
		if !foundImplementations[skillID] {
			return Report{}, fmt.Errorf("skill behavior coverage: declared skill %d is absent from mounted Skills.txt", skillID)
		}
	}

	result := Report{Schema: ReportSchema, Target: Target, Sources: sources}
	keys := sortedKeys(groups)
	for _, key := range keys {
		group := groups[key]
		sort.Slice(group.Consumers, func(i, j int) bool { return group.Consumers[i].SkillID < group.Consumers[j].SkillID })
		result.Behaviors = append(result.Behaviors, *group)
	}
	result.Summary = Summary{
		SkillRows: len(skills), BehaviorGroups: len(result.Behaviors),
		ImplementedSkills: len(foundImplementations), MissingSkills: len(skills) - len(foundImplementations),
	}
	return result, nil
}

func sortedKeys[T any](values map[string]T) []string {
	result := make([]string, 0, len(values))
	for value := range values {
		result = append(result, value)
	}
	sort.Strings(result)
	return result
}
