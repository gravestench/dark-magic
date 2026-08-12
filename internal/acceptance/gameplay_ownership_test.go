package acceptance

import (
	"bufio"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

type ownershipRule struct {
	prefix      string
	class       string
	destination string
}

// TestGameplayOwnershipInventoryIsExhaustive turns the migration inventory into
// a gate. A new gameplay or Lua-runtime file needs an explicit destination
// before it can quietly deepen the old Go/D2 coupling.
func TestGameplayOwnershipInventoryIsExhaustive(t *testing.T) {
	root := repositoryRoot(t)
	rules := readOwnershipRules(t, filepath.Join(root, "docs", "architecture", "gameplay-ownership.tsv"))
	roots := []string{
		"internal/game",
		"internal/runtime/lua",
		"internal/content/d2legacy/lua",
	}
	for _, relativeRoot := range roots {
		absoluteRoot := filepath.Join(root, filepath.FromSlash(relativeRoot))
		err := filepath.WalkDir(absoluteRoot, func(path string, entry os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if entry.IsDir() || !isProductionOwnershipFile(path) {
				return nil
			}
			relative, err := filepath.Rel(root, path)
			if err != nil {
				return err
			}
			relative = filepath.ToSlash(relative)
			matches := 0
			for _, rule := range rules {
				if strings.HasPrefix(relative, rule.prefix) {
					matches++
				}
			}
			if matches != 1 {
				t.Errorf("%s matches %d ownership rules; every production file needs exactly one destination", relative, matches)
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
}

// TestMigratedGameplayCoverageInventoryHasNoUnknownStatus makes replacement
// coverage a checked migration artifact. A deleted Go policy family cannot
// silently disappear: it must name its authoritative evidence and whether that
// evidence is complete, partial, transitional, or still pending.
func TestMigratedGameplayCoverageInventoryHasNoUnknownStatus(t *testing.T) {
	root := repositoryRoot(t)
	path := filepath.Join(root, "docs", "architecture", "d2legacy-test-coverage.tsv")
	file, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	seen := map[string]bool{}
	allowed := map[string]bool{
		"covered": true, "partial": true, "pending": true, "transitional": true,
	}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Split(line, "\t")
		if len(fields) != 4 {
			t.Fatalf("invalid coverage row %q", line)
		}
		if seen[fields[0]] {
			t.Errorf("duplicate coverage family %q", fields[0])
		}
		seen[fields[0]] = true
		if !allowed[fields[2]] {
			t.Errorf("coverage family %q has unknown status %q", fields[0], fields[2])
		}
		if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(fields[1]))); err != nil {
			t.Errorf("coverage evidence %q is missing: %v", fields[1], err)
		}
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
}

// TestD2LegacyLuaTestsStayReadable protects the bundled mod's role as living
// documentation. Syntax and behavior are covered by the suite runner; this
// gate catches the compact generated style that makes otherwise-correct Lua
// difficult for mod authors to learn from or maintain.
func TestD2LegacyLuaTestsStayReadable(t *testing.T) {
	root := repositoryRoot(t)
	luaRoot := filepath.Join(root, "internal", "content", "d2legacy", "lua", "d2legacy")
	err := filepath.WalkDir(luaRoot, func(path string, entry os.DirEntry, err error) error {
		if err != nil || entry.IsDir() {
			return err
		}
		relative, _ := filepath.Rel(luaRoot, path)
		relative = filepath.ToSlash(relative)
		isSuite := strings.HasSuffix(path, "_test.lua")
		isSupport := strings.HasPrefix(relative, "tests/") && strings.HasSuffix(path, ".lua")
		if !isSuite && !isSupport {
			return nil
		}
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()
		scanner := bufio.NewScanner(file)
		line := 0
		for scanner.Scan() {
			line++
			text := scanner.Text()
			if strings.ContainsRune(text, '\t') {
				t.Errorf("%s:%d uses a tab; d2legacy Lua uses four spaces", relative, line)
			}
			if len(text) > 120 {
				t.Errorf("%s:%d is %d columns; split fixtures and expressions for readability", relative, line, len(text))
			}
		}
		return scanner.Err()
	})
	if err != nil {
		t.Fatal(err)
	}
}

// TestD2LegacyLuaTestArchitecture keeps test ownership from drifting back into
// Go strings or private one-off Lua runners. Suites use the versioned authoring
// API and structured data so the host remains a generic production launcher.
func TestD2LegacyLuaTestArchitecture(t *testing.T) {
	root := repositoryRoot(t)
	luaRoot := filepath.Join(root, "internal", "content", "d2legacy", "lua", "d2legacy")
	err := filepath.WalkDir(luaRoot, func(path string, entry os.DirEntry, err error) error {
		if err != nil || entry.IsDir() || !strings.HasSuffix(path, ".lua") {
			return err
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		text := string(data)
		relative, _ := filepath.Rel(luaRoot, path)
		relative = filepath.ToSlash(relative)
		if strings.HasSuffix(path, "_test.lua") {
			for _, required := range []string{
				`require("d2legacy.tests/v1")`, "return test.suite(", "profile = ", "tier = ",
			} {
				if !strings.Contains(text, required) {
					t.Errorf("%s bypasses the versioned Lua test API; missing %q", relative, required)
				}
			}
			for _, forbidden := range []string{"initial_data_json", "records_json", "payload = [["} {
				if strings.Contains(text, forbidden) {
					t.Errorf("%s uses %q; use structured Lua tables so the harness owns serialization", relative, forbidden)
				}
			}
			return nil
		}
		if !strings.HasPrefix(relative, "tests/") && strings.Contains(text, "d2legacy.tests") {
			t.Errorf("production module %s imports test-only support", relative)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	goRoot := filepath.Join(root, "internal", "mod", "d2legacy")
	err = filepath.WalkDir(goRoot, func(path string, entry os.DirEntry, err error) error {
		if err != nil || entry.IsDir() || !strings.HasSuffix(path, "_test.go") {
			return err
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		for _, forbidden := range []string{"DoString(", "lua.DoString", "L.DoString"} {
			if strings.Contains(string(data), forbidden) {
				relative, _ := filepath.Rel(root, path)
				t.Errorf("%s embeds Lua through %q; put the scenario in a checked-in Lua fixture", filepath.ToSlash(relative), forbidden)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

// TestLuaNamespacesDescribeOwnership prevents the retired two-letter
// abbreviation from blurring generic engine APIs with the d2legacy mod again. Engine doors
// use engine.*; Diablo runtime state and bundled modules use d2legacy.*.
func TestLuaNamespacesDescribeOwnership(t *testing.T) {
	root := repositoryRoot(t)
	for _, relativeRoot := range []string{"internal", "cmd"} {
		err := filepath.WalkDir(filepath.Join(root, relativeRoot), func(path string, entry os.DirEntry, err error) error {
			if err != nil || entry.IsDir() {
				return err
			}
			if extension := filepath.Ext(path); extension != ".go" && extension != ".lua" {
				return nil
			}
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			text := string(data)
			retiredShort := "d" + "m."
			retiredModShort := "d" + "2."
			retiredLong := "dark" + "magic"
			// Only quoted d2.* identifiers are namespaces. Real legacy archive
			// names such as patch_d2.mpq are content filenames, not Lua APIs.
			usesRetiredModShort := strings.Contains(text, `"`+retiredModShort) || strings.Contains(text, `'`+retiredModShort)
			if strings.Contains(text, retiredShort) || usesRetiredModShort || strings.Contains(text, retiredLong+".") || strings.Contains(text, retiredLong+"/") {
				relative, _ := filepath.Rel(root, path)
				t.Errorf("%s uses a retired Lua namespace; use engine.* or d2legacy.*", filepath.ToSlash(relative))
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
}

// TestD2LegacyContentRetiresShimIdentity protects the mod boundary in active
// production code. Historical research may still discuss older architecture,
// but executable paths and identifiers must call the first-party mod d2legacy.
func TestD2LegacyContentRetiresShimIdentity(t *testing.T) {
	root := repositoryRoot(t)
	for _, relativeRoot := range []string{"internal/content", "internal/dev/tools/d2legacy_pack"} {
		err := filepath.WalkDir(filepath.Join(root, relativeRoot), func(path string, entry os.DirEntry, err error) error {
			if err != nil || entry.IsDir() || strings.HasSuffix(path, "_test.go") {
				return err
			}
			if extension := filepath.Ext(path); extension != ".go" && extension != ".lua" && extension != ".md" {
				return nil
			}
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			retired := "s" + "him"
			if strings.Contains(strings.ToLower(string(data)), retired) {
				relative, _ := filepath.Rel(root, path)
				t.Errorf("%s uses retired first-party content terminology; use d2legacy", filepath.ToSlash(relative))
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
}

// TestClientDoesNotReinstallMigratedD2Policy keeps completed Lua cutovers from
// becoming optional decorations over a second Go authority path.
func TestClientDoesNotReinstallMigratedD2Policy(t *testing.T) {
	root := repositoryRoot(t)
	for _, name := range []string{"core.go", "runtime.go", "components.go"} {
		path := filepath.Join(root, "internal", "app", "clientapp", name)
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		text := string(data)
		for _, forbidden := range []string{
			"RegisterPlayerBasicAttack", "RegisterAI", "RegisterDeath",
			"game/loot\"", "gamemonster.Register,", "gameplayer.Register,",
			"gametransition.Register(",
			"gamestate.Register(",
			"gameplayer.RegisterEquipmentProfile(",
		} {
			if strings.Contains(text, forbidden) {
				t.Errorf("%s reinstalls migrated d2legacy policy through %q", name, forbidden)
			}
		}
	}
}

// Cross-zone policy used to be split between a native proximity command source
// and Lua command application. The fixed-tick d2legacy system now owns the
// complete decision, so no production Go file may recreate those authority
// types or the retired command identity.
func TestWorldTransitionPolicyStaysInD2LegacyLua(t *testing.T) {
	root := repositoryRoot(t)
	for _, relativeRoot := range []string{"internal/app", "internal/mod/d2legacy"} {
		err := filepath.WalkDir(filepath.Join(root, filepath.FromSlash(relativeRoot)), func(path string, entry os.DirEntry, err error) error {
			if err != nil || entry.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
				return err
			}
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			for _, forbidden := range []string{
				"system.world.transition", "transitionAuthority", "transitionSource",
				"act1-town:exit-", "town-entry", "NewActOneTownMoorSeam",
			} {
				if strings.Contains(string(data), forbidden) {
					relative, _ := filepath.Rel(root, path)
					t.Errorf("%s restores native D2 transition policy through %q", filepath.ToSlash(relative), forbidden)
				}
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
}

func TestD2MapTilePolicyStaysInD2LegacyLua(t *testing.T) {
	root := repositoryRoot(t)
	for _, relativeRoot := range []string{"internal/game", "internal/mod/d2legacy/adapter"} {
		err := filepath.WalkDir(filepath.Join(root, filepath.FromSlash(relativeRoot)), func(path string, entry os.DirEntry, err error) error {
			if err != nil || entry.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
				return err
			}
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			for _, forbidden := range []string{"actOneDirtPathSequence", "RealizeActOneDirtPath", "ActOneTownEntry", "ObjectTypeStatic && object.ID == 2"} {
				if strings.Contains(string(data), forbidden) {
					relative, _ := filepath.Rel(root, path)
					t.Errorf("%s restores native D2 map-tile policy through %q", filepath.ToSlash(relative), forbidden)
				}
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
}

// TestGameplayMechanismsDoNotGainPolicyDependencies is a migration ratchet.
// The small debt file names today's known violations; no new generic engine
// mechanism may import a package classified as D2 policy or transitional.
func TestGameplayMechanismsDoNotGainPolicyDependencies(t *testing.T) {
	root := repositoryRoot(t)
	rules := readOwnershipRules(t, filepath.Join(root, "docs", "architecture", "gameplay-ownership.tsv"))
	debt := readDependencyDebt(t, filepath.Join(root, "docs", "architecture", "gameplay-dependency-debt.tsv"))
	const projectPrefix = "github.com/gravestench/dark-magic/"

	for _, rule := range rules {
		if rule.class != "mechanism" {
			continue
		}
		mechanismRoot := filepath.Join(root, filepath.FromSlash(strings.TrimSuffix(rule.prefix, "/")))
		err := filepath.WalkDir(mechanismRoot, func(path string, entry os.DirEntry, err error) error {
			if err != nil || entry.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
				return err
			}
			file, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.ImportsOnly)
			if err != nil {
				return err
			}
			importer, err := filepath.Rel(root, path)
			if err != nil {
				return err
			}
			importer = filepath.ToSlash(importer)
			for _, imported := range file.Imports {
				name, err := strconv.Unquote(imported.Path.Value)
				if err != nil || !strings.HasPrefix(name, projectPrefix) {
					continue
				}
				dependency := strings.TrimPrefix(name, projectPrefix)
				dependencyClass := ownershipClassForPath(rules, dependency+"/")
				if dependencyClass != "d2-policy" && dependencyClass != "transitional" {
					continue
				}
				edge := importer + "\t" + dependency
				if _, allowed := debt[edge]; !allowed {
					t.Errorf("%s imports %s (%s); generic mechanisms may not gain D2 policy dependencies", importer, dependency, dependencyClass)
				}
				delete(debt, edge)
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	for edge := range debt {
		t.Errorf("stale gameplay dependency debt %s; remove the resolved exception", edge)
	}
}

// TestGenericEngineDoesNotImportFirstPartyMod is the hard architectural wall:
// engine packages may expose generic capabilities used by d2legacy, but the
// dependency arrow can never point back from the engine into the bundled mod.
func TestGenericEngineDoesNotImportFirstPartyMod(t *testing.T) {
	root := repositoryRoot(t)
	const projectPrefix = "github.com/gravestench/dark-magic/"
	for _, relativeRoot := range []string{"internal/game", "internal/runtime/lua"} {
		err := filepath.WalkDir(filepath.Join(root, filepath.FromSlash(relativeRoot)), func(path string, entry os.DirEntry, err error) error {
			if err != nil || entry.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
				return err
			}
			file, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.ImportsOnly)
			if err != nil {
				return err
			}
			for _, imported := range file.Imports {
				name, err := strconv.Unquote(imported.Path.Value)
				if err == nil && strings.HasPrefix(name, projectPrefix+"internal/mod/d2legacy") {
					relative, _ := filepath.Rel(root, path)
					t.Errorf("%s imports first-party mod package %s", filepath.ToSlash(relative), name)
				}
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
}

// TestGenericEngineDoesNotNameD2Policy catches the subtler form of coupling
// where rules are copied into an engine package without importing d2legacy.
// Typed record schemas and codec-facing world data are deliberately outside
// this lexical gate: they may describe legacy data, but must not decide what
// that data means during gameplay.
func TestGenericEngineDoesNotNameD2Policy(t *testing.T) {
	root := repositoryRoot(t)
	forbidden := []string{
		"d2legacy", "fire bolt", "fire_bolt", "treasure class", "horadric",
		"rogue encampment", "blood moor", "monstats", "amazon", "sorceress",
		"necromancer", "barbarian", "paladin", "druid", "assassin",
	}
	for _, relativeRoot := range []string{"internal/game/ecs", "internal/game/session", "internal/game/simulation", "internal/game/worldgen", "internal/runtime/lua"} {
		err := filepath.WalkDir(filepath.Join(root, filepath.FromSlash(relativeRoot)), func(path string, entry os.DirEntry, err error) error {
			if err != nil || entry.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
				return err
			}
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			lower := strings.ToLower(string(data))
			for _, token := range forbidden {
				if strings.Contains(lower, token) {
					relative, _ := filepath.Rel(root, path)
					t.Errorf("%s names Diablo policy %q; generic engine code must use mod-neutral vocabulary", filepath.ToSlash(relative), token)
				}
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
}

func readOwnershipRules(t *testing.T, path string) []ownershipRule {
	t.Helper()
	file, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	var rules []ownershipRule
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Split(line, "\t")
		if len(fields) != 4 {
			t.Fatalf("%s contains malformed rule %q", path, line)
		}
		rules = append(rules, ownershipRule{prefix: fields[0], class: fields[1], destination: fields[2]})
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
	if len(rules) == 0 {
		t.Fatal("gameplay ownership inventory has no rules")
	}
	return rules
}

func readDependencyDebt(t *testing.T, path string) map[string]struct{} {
	t.Helper()
	file, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	debt := make(map[string]struct{})
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Split(line, "\t")
		if len(fields) != 3 {
			t.Fatalf("%s contains malformed debt %q", path, line)
		}
		debt[fields[0]+"\t"+fields[1]] = struct{}{}
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
	return debt
}

func ownershipClassForPath(rules []ownershipRule, path string) string {
	for _, rule := range rules {
		if strings.HasPrefix(path, rule.prefix) {
			return rule.class
		}
	}
	return ""
}

func isProductionOwnershipFile(path string) bool {
	if strings.HasSuffix(path, "_test.go") || strings.HasSuffix(path, "_test.lua") {
		return false
	}
	if strings.Contains(filepath.ToSlash(path), "/lua/d2legacy/tests/") {
		return false
	}
	return strings.HasSuffix(path, ".go") || strings.HasSuffix(path, ".lua")
}
