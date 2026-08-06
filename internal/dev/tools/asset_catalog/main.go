package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/gravestench/dark-magic/internal/assets/catalog"
	"github.com/gravestench/dark-magic/internal/content"
	darkpaths "github.com/gravestench/dark-magic/internal/paths"
)

func main() {
	mpqDirectory := flag.String("mpq-dir", os.Getenv("MPQ_DIRECTORY"), "directory containing Diablo II MPQ files")
	manifestPath := flag.String("manifest", "", "optional JSON manifest; defaults to the curated screen catalog")
	listfilePath := flag.String("listfile", "", "optional community MPQ listfile to audit against this installation")
	outputDirectory := flag.String("out", "asset-catalog", "output directory for report.json and contact sheets")
	noSheets := flag.Bool("no-sheets", false, "skip DC6 contact sheet generation")
	writeFixture := flag.String("write-fixture", "", "write a structural fixture after every manifest asset verifies")
	fixturePath := flag.String("fixture", "", "validate the installation against a structural fixture")
	flag.Parse()

	if *mpqDirectory == "" {
		fatal("-mpq-dir or MPQ_DIRECTORY is required")
	}
	expandedMPQDirectory, err := darkpaths.ExpandHost(*mpqDirectory)
	if err != nil {
		fatal(err.Error())
	}
	expandedManifestPath, err := darkpaths.ExpandHost(*manifestPath)
	if err != nil {
		fatal(err.Error())
	}
	expandedOutputDirectory, err := darkpaths.ExpandHost(*outputDirectory)
	if err != nil {
		fatal(err.Error())
	}
	expandedListfilePath, err := darkpaths.ExpandHost(*listfilePath)
	if err != nil {
		fatal(err.Error())
	}
	expandedWriteFixture, err := darkpaths.ExpandHost(*writeFixture)
	if err != nil {
		fatal(err.Error())
	}
	expandedFixturePath, err := darkpaths.ExpandHost(*fixturePath)
	if err != nil {
		fatal(err.Error())
	}
	if expandedWriteFixture != "" && expandedFixturePath != "" {
		fatal("-write-fixture and -fixture are mutually exclusive")
	}
	manifest, err := loadManifest(expandedManifestPath)
	if err != nil {
		fatal(err.Error())
	}
	if err := manifest.Validate(); err != nil {
		fatal(err.Error())
	}

	contentFS, err := openMPQStack(expandedMPQDirectory)
	if err != nil {
		fatal(err.Error())
	}
	if err := os.MkdirAll(expandedOutputDirectory, 0o755); err != nil {
		fatal(err.Error())
	}

	options := assetcatalog.Options{Resolve: func(name string) (assetcatalog.Source, error) {
		source, err := contentFS.Resolve(name)
		return assetcatalog.Source{Layer: source.Layer, Path: source.Path}, err
	}}
	if !*noSheets {
		sheetsDirectory := filepath.Join(expandedOutputDirectory, "contact-sheets")
		if err := os.MkdirAll(sheetsDirectory, 0o755); err != nil {
			fatal(err.Error())
		}
		options.SheetWriter = func(name string, data []byte) (string, error) {
			path := filepath.Join(sheetsDirectory, name)
			if err := os.WriteFile(path, data, 0o644); err != nil {
				return "", err
			}
			return filepath.ToSlash(filepath.Join("contact-sheets", name)), nil
		}
	}

	report := assetcatalog.Verify(contentFS, manifest, options)
	reportPath := filepath.Join(expandedOutputDirectory, "report.json")
	file, err := os.Create(reportPath)
	if err != nil {
		fatal(err.Error())
	}
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	encodeErr := encoder.Encode(report)
	closeErr := file.Close()
	if encodeErr != nil {
		fatal(encodeErr.Error())
	}
	if closeErr != nil {
		fatal(closeErr.Error())
	}

	found := 0
	for _, result := range report.Results {
		if result.Found {
			found++
		}
	}
	fmt.Printf("verified %d/%d hypotheses; report: %s\n", found, len(report.Results), reportPath)
	if expandedWriteFixture != "" {
		fixture, err := assetcatalog.FixtureFromReport(report)
		if err != nil {
			fatal(err.Error())
		}
		if err := writeJSON(expandedWriteFixture, fixture); err != nil {
			fatal(err.Error())
		}
		fmt.Printf("wrote structural fixture: %s\n", expandedWriteFixture)
	}
	if expandedFixturePath != "" {
		var fixture assetcatalog.Fixture
		if err := readJSON(expandedFixturePath, &fixture); err != nil {
			fatal(err.Error())
		}
		if mismatches := assetcatalog.CompareFixture(report, fixture); len(mismatches) != 0 {
			fatal("fixture mismatch:\n  " + strings.Join(mismatches, "\n  "))
		}
		fmt.Printf("fixture verified: %s\n", expandedFixturePath)
	}
	if expandedListfilePath != "" {
		listfile, err := os.Open(expandedListfilePath)
		if err != nil {
			fatal(err.Error())
		}
		entries, parseErr := assetcatalog.ParseListfile(listfile)
		closeErr := listfile.Close()
		if parseErr != nil {
			fatal(parseErr.Error())
		}
		if closeErr != nil {
			fatal(closeErr.Error())
		}
		audit := assetcatalog.AuditListfile(contentFS, entries)
		auditPath := filepath.Join(expandedOutputDirectory, "listfile-report.json")
		output, err := os.Create(auditPath)
		if err != nil {
			fatal(err.Error())
		}
		encoder := json.NewEncoder(output)
		encoder.SetIndent("", "  ")
		encodeErr, closeErr := encoder.Encode(audit), output.Close()
		if encodeErr != nil {
			fatal(encodeErr.Error())
		}
		if closeErr != nil {
			fatal(closeErr.Error())
		}
		fmt.Printf("resolved %d/%d listed paths; report: %s\n", audit.Found, audit.Listed, auditPath)
	}
}

func readJSON(name string, destination any) error {
	data, err := os.ReadFile(name)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, destination)
}

func writeJSON(name string, value any) error {
	if err := os.MkdirAll(filepath.Dir(name), 0o755); err != nil {
		return err
	}
	file, err := os.Create(name)
	if err != nil {
		return err
	}
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	encodeErr, closeErr := encoder.Encode(value), file.Close()
	if encodeErr != nil {
		return encodeErr
	}
	return closeErr
}

func loadManifest(name string) (assetcatalog.Manifest, error) {
	var data []byte
	var err error
	if name == "" {
		data, err = fs.ReadFile(content.Shim(), "manifests/asset-catalog.v1.json")
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

func openMPQStack(directory string) (*content.FS, error) {
	priority := []string{
		"patch_d2.mpq", "d2exp.mpq", "d2data.mpq", "d2char.mpq",
		"d2music.mpq", "d2sfx.mpq", "d2speech.mpq", "d2video.mpq",
		"d2xmusic.mpq", "d2xtalk.mpq", "d2xvideo.mpq",
	}
	layers := make([]content.Layer, 0, len(priority))
	for _, name := range priority {
		path := filepath.Join(directory, name)
		if _, err := os.Stat(path); err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, err
		}
		archive, err := content.OpenSource(path)
		if err != nil {
			return nil, err
		}
		layers = append(layers, content.Layer{Name: name, FS: archive})
	}
	if len(layers) == 0 {
		return nil, fmt.Errorf("no supported MPQs found in %q", directory)
	}
	return content.New(layers...)
}

func fatal(message string) {
	message = strings.TrimSpace(message)
	if message == "" {
		message = fs.ErrInvalid.Error()
	}
	fmt.Fprintln(os.Stderr, message)
	os.Exit(1)
}
