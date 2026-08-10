package acceptance

import (
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"
)

func TestCaptureAllCatalogMatchesRegisteredScenes(t *testing.T) {
	root := repositoryRoot(t)
	makefile, err := os.ReadFile(filepath.Join(root, "Makefile"))
	if err != nil {
		t.Fatal(err)
	}
	boot, err := os.ReadFile(filepath.Join(root, "internal/content/shim/boot.lua"))
	if err != nil {
		t.Fatal(err)
	}

	captureMatch := regexp.MustCompile(`(?m)^CAPTURE_ALL_SCENES := (.+)$`).FindSubmatch(makefile)
	if len(captureMatch) != 2 {
		t.Fatal("Makefile has no single-line CAPTURE_ALL_SCENES catalog")
	}
	captured := strings.Fields(string(captureMatch[1]))
	registered := make([]string, 0, len(captured))
	for _, match := range regexp.MustCompile(`scenes\.register\("([a-z0-9_]+)"`).FindAllSubmatch(boot, -1) {
		registered = append(registered, string(match[1]))
	}
	for _, match := range regexp.MustCompile(`(?m)^\s+([a-z0-9_]+)=\{title=`).FindAllSubmatch(boot, -1) {
		registered = append(registered, string(match[1]))
	}
	slices.Sort(captured)
	slices.Sort(registered)
	if !slices.Equal(captured, registered) {
		t.Fatalf("capture scenes do not match boot registrations\ncapture: %s\nregistered: %s", strings.Join(captured, ","), strings.Join(registered, ","))
	}
}
