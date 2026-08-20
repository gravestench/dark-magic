package ecs

import (
	"fmt"
	"reflect"
)

// FirstDifference describes the first canonical snapshot divergence in serialization order. It returns an empty string
// only when every replay-relevant field is deeply equal.
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
		difference := firstComponentDifference(
			componentIndex,
			expected.Components[componentIndex],
			actual.Components[componentIndex],
		)
		if difference != "" {
			return difference
		}
	}

	return ""
}

// firstComponentDifference compares schema metadata before instances, matching the serialized structure's hierarchy so
// callers receive the earliest actionable divergence instead of a downstream value mismatch.
func firstComponentDifference(index int, expected, actual ComponentSnapshot) string {
	if expected.Name != actual.Name {
		return fmt.Sprintf("component[%d] name %q != %q", index, expected.Name, actual.Name)
	}

	if expected.Version != actual.Version {
		return fmt.Sprintf("component %q version %d != %d", expected.Name, expected.Version, actual.Version)
	}

	if !reflect.DeepEqual(expected.Fields, actual.Fields) {
		return fmt.Sprintf("component %q schema differs", expected.Name)
	}

	if len(expected.Instances) != len(actual.Instances) {
		return fmt.Sprintf(
			"component %q instance count %d != %d",
			expected.Name,
			len(expected.Instances),
			len(actual.Instances),
		)
	}

	for instanceIndex := range expected.Instances {
		difference := firstInstanceDifference(
			expected.Name,
			expected.Fields,
			instanceIndex,
			expected.Instances[instanceIndex],
			actual.Instances[instanceIndex],
		)
		if difference != "" {
			return difference
		}
	}

	return ""
}

// firstInstanceDifference validates identity and shape before field values. This ordering distinguishes malformed
// snapshots from ordinary state divergence and avoids indexing a field list whose length does not match the values.
func firstInstanceDifference(
	componentName string,
	fields []FieldSnapshot,
	index int,
	expected InstanceSnapshot,
	actual InstanceSnapshot,
) string {
	if expected.Entity != actual.Entity {
		return fmt.Sprintf(
			"component %q entity[%d] %d != %d",
			componentName,
			index,
			expected.Entity,
			actual.Entity,
		)
	}

	if len(expected.Values) != len(actual.Values) {
		return fmt.Sprintf(
			"component %q entity %d value count %d != %d",
			componentName,
			expected.Entity,
			len(expected.Values),
			len(actual.Values),
		)
	}

	if len(expected.Values) != len(fields) {
		return fmt.Sprintf(
			"component %q entity %d has %d values for %d fields",
			componentName,
			expected.Entity,
			len(expected.Values),
			len(fields),
		)
	}

	for fieldIndex := range expected.Values {
		if !reflect.DeepEqual(expected.Values[fieldIndex], actual.Values[fieldIndex]) {
			return fmt.Sprintf(
				"component %q entity %d field %q differs",
				componentName,
				expected.Entity,
				fields[fieldIndex].Name,
			)
		}
	}

	return ""
}
