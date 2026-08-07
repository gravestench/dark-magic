package ecs

import (
	"fmt"
	"reflect"
)

// FirstDifference describes the first canonical snapshot divergence.
func FirstDifference(expected, actual Snapshot) string {
	if expected.Version != actual.Version {
		return fmt.Sprintf("version %d != %d", expected.Version, actual.Version)
	}
	if expected.Tick != actual.Tick {
		return fmt.Sprintf("tick %d != %d", expected.Tick, actual.Tick)
	}
	if !reflect.DeepEqual(expected.Entities, actual.Entities) {
		return fmt.Sprintf("entities %v != %v", expected.Entities, actual.Entities)
	}
	if len(expected.Components) != len(actual.Components) {
		return fmt.Sprintf("component count %d != %d", len(expected.Components), len(actual.Components))
	}
	for componentIndex := range expected.Components {
		e, a := expected.Components[componentIndex], actual.Components[componentIndex]
		if e.Name != a.Name {
			return fmt.Sprintf("component[%d] name %q != %q", componentIndex, e.Name, a.Name)
		}
		if e.Version != a.Version {
			return fmt.Sprintf("component %q version %d != %d", e.Name, e.Version, a.Version)
		}
		if !reflect.DeepEqual(e.Fields, a.Fields) {
			return fmt.Sprintf("component %q schema differs", e.Name)
		}
		if len(e.Instances) != len(a.Instances) {
			return fmt.Sprintf("component %q instance count %d != %d", e.Name, len(e.Instances), len(a.Instances))
		}
		for instanceIndex := range e.Instances {
			ei, ai := e.Instances[instanceIndex], a.Instances[instanceIndex]
			if ei.Entity != ai.Entity {
				return fmt.Sprintf("component %q entity[%d] %d != %d", e.Name, instanceIndex, ei.Entity, ai.Entity)
			}
			if len(ei.Values) != len(ai.Values) {
				return fmt.Sprintf("component %q entity %d value count %d != %d", e.Name, ei.Entity, len(ei.Values), len(ai.Values))
			}
			if len(ei.Values) != len(e.Fields) {
				return fmt.Sprintf("component %q entity %d has %d values for %d fields", e.Name, ei.Entity, len(ei.Values), len(e.Fields))
			}
			for fieldIndex := range ei.Values {
				if !reflect.DeepEqual(ei.Values[fieldIndex], ai.Values[fieldIndex]) {
					return fmt.Sprintf("component %q entity %d field %q differs", e.Name, ei.Entity, e.Fields[fieldIndex].Name)
				}
			}
		}
	}
	return ""
}
