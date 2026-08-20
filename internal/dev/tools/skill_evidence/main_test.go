package main

import (
	"reflect"
	"testing"
)

// TestParseSkillIDsPreservesRequestedSequence protects duplicate and ordering semantics in generated evidence.
func TestParseSkillIDsPreservesRequestedSequence(t *testing.T) {
	t.Parallel()

	got, err := parseSkillIDs("40, 0,40")
	if err != nil {
		t.Fatalf("parseSkillIDs() error = %v", err)
	}

	want := []int{40, 0, 40}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("parseSkillIDs() = %v, want %v", got, want)
	}
}

// TestParseSkillIDsRejectsInvalidValues protects the command from silently selecting unintended skills.
func TestParseSkillIDsRejectsInvalidValues(t *testing.T) {
	t.Parallel()

	for _, value := range []string{"unknown", "-1", ""} {
		value := value
		t.Run(value, func(t *testing.T) {
			t.Parallel()

			if _, err := parseSkillIDs(value); err == nil {
				t.Fatalf("parseSkillIDs(%q) unexpectedly succeeded", value)
			}
		})
	}
}
