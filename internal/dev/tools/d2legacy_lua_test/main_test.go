package main

import "testing"

func TestModuleName(t *testing.T) {
	got, ok := moduleName("internal/content/d2legacy/lua/d2legacy/policy/damage.lua")
	if !ok || got != "d2legacy.policy.damage" {
		t.Fatalf("moduleName() = %q, %v", got, ok)
	}
}
