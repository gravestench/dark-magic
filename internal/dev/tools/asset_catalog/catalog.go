package main

import (
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"

	assetcatalog "github.com/gravestench/dark-magic/internal/assets/catalog"
	"github.com/gravestench/dark-magic/internal/content"
	darkpaths "github.com/gravestench/dark-magic/internal/paths"
)

// expandHostPath applies the repository's cross-platform host-path rules. Keeping that policy behind this package
// boundary prevents virtual MPQ paths from being expanded accidentally.
func expandHostPath(name string) (string, error) {
	return darkpaths.ExpandHost(name)
}

// verifyCatalog prepares the manifest, archive stack, and output writers before starting the scan. Verification only
// begins after every prerequisite succeeds, so configuration failures cannot leave a partial report.
func verifyCatalog(options commandOptions) (*content.FS, assetcatalog.Report, error) {
	manifest, err := loadManifest(options.manifestPath)
	if err != nil {
		return nil, assetcatalog.Report{}, err
	}

	if err := manifest.Validate(); err != nil {
		return nil, assetcatalog.Report{}, err
	}

	contentFS, err := openMPQStack(options.mpqDirectory)
	if err != nil {
		return nil, assetcatalog.Report{}, err
	}

	if err := os.MkdirAll(options.outputDirectory, 0o755); err != nil {
		return nil, assetcatalog.Report{}, err
	}

	verificationOptions, err := catalogVerificationOptions(contentFS, options.outputDirectory, options.noSheets)
	if err != nil {
		return nil, assetcatalog.Report{}, err
	}

	return contentFS, assetcatalog.Verify(contentFS, manifest, verificationOptions), nil
}

// loadManifest reads either the embedded curated catalog or an explicit host file. Both sources share the same JSON
// decoding path so selecting a source cannot alter the manifest contract.
func loadManifest(name string) (assetcatalog.Manifest, error) {
	var (
		data []byte
		err  error
	)
	if name == "" {
		data, err = fs.ReadFile(content.D2Legacy(), "manifests/asset-catalog.v1.json")
	} else {
		data, err = os.ReadFile(name)
	}

	if err != nil {
		return assetcatalog.Manifest{}, err
	}

	var manifest assetcatalog.Manifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return assetcatalog.Manifest{}, err
	}

	return manifest, nil
}

// catalogVerificationOptions binds catalog diagnostics to the selected MPQ stack and output directory. Contact-sheet
// storage is prepared eagerly so Verify never begins when its optional writer cannot be configured.
func catalogVerificationOptions(
	contentFS *content.FS,
	outputDirectory string,
	noSheets bool,
) (assetcatalog.Options, error) {
	options := assetcatalog.Options{Resolve: catalogSourceResolver(contentFS)}
	if noSheets {
		return options, nil
	}

	sheetsDirectory := filepath.Join(outputDirectory, "contact-sheets")
	if err := os.MkdirAll(sheetsDirectory, 0o755); err != nil {
		return assetcatalog.Options{}, err
	}

	options.SheetWriter = contactSheetWriter(sheetsDirectory)

	return options, nil
}

// catalogSourceResolver exposes only stable layer and path provenance, keeping the catalog package independent from
// the richer content resolver while preserving resolution errors unchanged.
func catalogSourceResolver(contentFS *content.FS) func(string) (assetcatalog.Source, error) {
	// Return the resolver error unchanged while translating only the provenance fields the report contract exposes.
	return func(name string) (assetcatalog.Source, error) {
		source, err := contentFS.Resolve(name)
		return assetcatalog.Source{Layer: source.Layer, Path: source.Path}, err
	}
}

// contactSheetWriter stores generated sheets beneath the report directory and returns slash-separated relative paths
// so report JSON stays portable across host operating systems.
func contactSheetWriter(sheetsDirectory string) func(string, []byte) (string, error) {
	// Publish a relative report path only after the owned artifact has been written successfully.
	return func(name string, data []byte) (string, error) {
		path := filepath.Join(sheetsDirectory, name)
		if err := os.WriteFile(path, data, 0o644); err != nil {
			return "", err
		}

		return filepath.ToSlash(filepath.Join("contact-sheets", name)), nil
	}
}

// foundHypothesisCount derives the human-facing summary from result order without mutating the deterministic report.
func foundHypothesisCount(report assetcatalog.Report) int {
	found := 0

	for _, result := range report.Results {
		if result.Found {
			found++
		}
	}

	return found
}
