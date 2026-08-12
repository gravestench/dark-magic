// Package presentation adapts d2legacy's authored presentation documents at
// the client composition boundary. The generic data capability knows only how
// to validate and decode manifests; it never names this mod's schema.
package presentation

import (
	"github.com/gravestench/dark-magic/internal/content"
	modruntime "github.com/gravestench/dark-magic/internal/runtime/lua"
)

const Schema = "d2legacy.presentation/v1"

func ManifestTransforms(profile string) map[string]modruntime.ManifestTransform {
	return map[string]modruntime.ManifestTransform{
		Schema: func(document map[string]any) (map[string]any, error) {
			result, _, err := content.ApplyPresentationProfile(document, profile)
			return result, err
		},
	}
}
