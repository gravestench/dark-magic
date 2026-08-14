package clientapp

import (
	"testing"

	"github.com/gravestench/dark-magic/internal/modcache"
)

func TestModlessStartupIsExplicitAndDoesNotTreatLegacyTestOptionsAsEmpty(t *testing.T) {
	if !(&application{options: Options{Mods: &modcache.Lock{Schema: modcache.LockSchema}}}).modless() {
		t.Fatal("explicit empty resolved mod set did not select mod-neutral startup")
	}
	if (&application{options: Options{}}).modless() {
		t.Fatal("unspecified test composition unexpectedly selected mod-neutral startup")
	}
	lock := &modcache.Lock{Schema: modcache.LockSchema, Packages: []modcache.LockedPackage{{Manifest: modcache.Manifest{ID: "example"}}}}
	if (&application{options: Options{Mods: lock}}).modless() {
		t.Fatal("non-empty resolved mod set selected mod-neutral startup")
	}
}
