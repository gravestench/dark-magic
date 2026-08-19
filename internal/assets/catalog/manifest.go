package assetcatalog

import (
	"errors"
	"fmt"
)

// Validate rejects manifests whose identity and lookup fields cannot support deterministic verification.
// Validation remains independent of game data so callers can fail before opening proprietary archives.
func (m Manifest) Validate() error {
	if m.Version < 1 {
		return errors.New("asset catalog: manifest version must be positive")
	}

	seen := make(map[string]struct{}, len(m.Assets))
	for index, asset := range m.Assets {
		if asset.ID == "" || asset.Screen == "" || asset.Path == "" || asset.Meaning == "" {
			return fmt.Errorf("asset catalog: asset %d requires id, screen, path, and meaning", index)
		}

		if _, exists := seen[asset.ID]; exists {
			return fmt.Errorf("asset catalog: duplicate id %q", asset.ID)
		}

		seen[asset.ID] = struct{}{}
	}

	return nil
}
