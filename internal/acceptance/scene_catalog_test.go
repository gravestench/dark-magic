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
	bootstrap := readBootstrapLua(t, root)

	captureMatch := regexp.MustCompile(`(?m)^CAPTURE_ALL_SCENES := (.+)$`).FindSubmatch(makefile)
	if len(captureMatch) != 2 {
		t.Fatal("Makefile has no single-line CAPTURE_ALL_SCENES catalog")
	}
	captured := strings.Fields(string(captureMatch[1]))
	registered := make([]string, 0, len(captured))
	for _, match := range regexp.MustCompile(`scenes\.register\("([a-z0-9_]+)"`).FindAllStringSubmatch(bootstrap, -1) {
		registered = append(registered, string(match[1]))
	}
	for _, match := range regexp.MustCompile(`(?m)^\s+([a-z0-9_]+)=\{(?:title|module)=`).FindAllStringSubmatch(bootstrap, -1) {
		registered = append(registered, string(match[1]))
	}
	// The plain overlays are intentionally one compact list because they need no
	// routing metadata. Pull those quoted names from that list too.
	plain := regexp.MustCompile(`ipairs\(\{([^}]+)\}\)`).FindStringSubmatch(bootstrap)
	if len(plain) == 2 {
		for _, match := range regexp.MustCompile(`"([a-z0-9_]+)"`).FindAllStringSubmatch(plain[1], -1) {
			registered = append(registered, match[1])
		}
	}
	slices.Sort(captured)
	slices.Sort(registered)
	if !slices.Equal(captured, registered) {
		t.Fatalf("capture scenes do not match boot registrations\ncapture: %s\nregistered: %s", strings.Join(captured, ","), strings.Join(registered, ","))
	}
}

func TestMessagesShellLeavesGameplayLive(t *testing.T) {
	t.Parallel()

	root := repositoryRoot(t)
	bootstrap := readBootstrapLua(t, root)
	definition := regexp.MustCompile(`messages=\{[^\n]+\}`).FindString(bootstrap)
	if definition == "" {
		t.Fatal("bootstrap modules have no messages shell definition")
	}
	want := []string{
		"blocks_update_below=false",
		"passes_input_below=true",
		`world_view="center"`,
		`layer="hud"`,
	}
	for _, fact := range want {
		if !strings.Contains(definition, fact) {
			t.Errorf("messages shell definition %q does not contain %q", definition, fact)
		}
	}
}

func TestGameWorldUsesChunkedAuthoritativeCameraAdapter(t *testing.T) {
	t.Parallel()

	root := repositoryRoot(t)
	path := filepath.Join(root, "internal/content/d2legacy/lua/d2legacy/screens/game_world.lua")
	contents, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	source := string(contents)
	for _, required := range []string{
		`require("d2legacy.presentation.chunked_map")`,
		"chunked_map.create(",
		"chunked_map.update(",
	} {
		if !strings.Contains(source, required) {
			t.Errorf("game_world is missing %q", required)
		}
	}
	for _, retired := range []string{"set_ds1(", "initial_camera_x", "initial_camera_y"} {
		if strings.Contains(source, retired) {
			t.Errorf("game_world still contains retired full-map/baseline path %q", retired)
		}
	}
}

func TestMapgenLabRegenerationUsesDocumentedRenderNodeLifetime(t *testing.T) {
	t.Parallel()

	root := repositoryRoot(t)
	path := filepath.Join(root, "internal/content/d2legacy/lua/d2legacy/screens/mapgen_lab.lua")
	contents, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	source := string(contents)
	if strings.Contains(source, "node:exists()") {
		t.Fatal("Mapgen Lab calls an exists method that engine.render/v1 nodes do not expose")
	}
	if !strings.Contains(source, "node:destroy()") {
		t.Fatal("Mapgen Lab does not release topology nodes before drawing another seed")
	}
}

func TestDS1LabUsesNativeChunkDepthForLayerOrdering(t *testing.T) {
	t.Parallel()

	root := repositoryRoot(t)
	path := filepath.Join(root, "internal/content/d2legacy/lua/d2legacy/screens/ds1_lab.lua")
	contents, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	source := string(contents)
	if !strings.Contains(source, "node:set_z(chunk.depth)") {
		t.Fatal("DS1 Lab does not use the native renderer's authoritative chunk depth")
	}
	if strings.Contains(source, "node:set_z(chunk.layer)") {
		t.Fatal("DS1 Lab uses semantic layer identities as draw-order values")
	}
}

func readBootstrapLua(t *testing.T, root string) string {
	t.Helper()
	paths, err := filepath.Glob(filepath.Join(root, "internal/content/d2legacy/lua/d2legacy/bootstrap/*.lua"))
	if err != nil {
		t.Fatal(err)
	}
	var source strings.Builder
	for _, path := range paths {
		contents, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		source.Write(contents)
		source.WriteByte('\n')
	}
	return source.String()
}
