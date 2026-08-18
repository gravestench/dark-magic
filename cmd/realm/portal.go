package main

import (
	"fmt"
	"os"
	"path/filepath"

	portalassets "github.com/gravestench/dark-magic/internal/app/realmportal/assets"
	"github.com/gravestench/dark-magic/internal/content"
)

// preparePortalAssets builds the optional game-art cache used by the Realm portal.
func preparePortalAssets(directory string) (*portalassets.Cache, func(), error) {
	if os.Getenv("MPQ_DIRECTORY") == "" {
		return nil, func() {}, nil
	}
	contentFS, err := content.FromEnvironment()
	if err != nil {
		return nil, nil, fmt.Errorf("open Realm portal game assets: %w", err)
	}
	closeContent := func() {
		_ = contentFS.Close()
	}
	assets, err := portalassets.New(
		contentFS,
		filepath.Join(directory, "cache", "realm-portal"),
	)
	if err != nil {
		closeContent()
		return nil, nil, fmt.Errorf("configure Realm portal asset cache: %w", err)
	}
	return assets, closeContent, nil
}
