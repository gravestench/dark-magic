// Package presentation adapts d2legacy's authored presentation documents at
// the client composition boundary. The generic data capability knows only how
// to validate and decode manifests; it never names this mod's schema.
package presentation

import (
	"github.com/gravestench/dark-magic/internal/content"
	modruntime "github.com/gravestench/dark-magic/internal/runtime/lua"
)

// Schema identifies authored presentation documents handled by this adapter.
const Schema = "d2legacy.presentation/v1"

// ManifestTransforms binds the d2legacy schema to one presentation profile selected at client composition time.
// Returning a fresh map prevents callers from sharing mutable transform registration across runtime instances.
func ManifestTransforms(profile string) map[string]modruntime.ManifestTransform {
	return map[string]modruntime.ManifestTransform{
		Schema: profileTransform(profile),
	}
}

// profileTransform captures the immutable profile while discarding metadata that is outside the manifest contract.
func profileTransform(profile string) modruntime.ManifestTransform {
	// Each closure owns only an immutable string, so runtime instances cannot share mutable adapter state.
	return func(document map[string]any) (map[string]any, error) {
		result, _, err := content.ApplyPresentationProfile(document, profile)
		return result, err
	}
}
