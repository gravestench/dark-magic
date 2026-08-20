package acceptance

import (
	"bufio"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// ownershipRule maps a production path prefix to its current architectural class and intended
// destination, turning the migration inventory into executable documentation.
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

	defer func() {
		if err := file.Close(); err != nil {
			t.Errorf("close coverage inventory: %v", err)
		}
	}()

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

		evidence := strings.SplitN(fields[1], "#", 2)
		if len(evidence) != 2 || evidence[1] == "" {
			t.Errorf("coverage family %q must name evidence as path#case-or-test", fields[0])
			continue
		}

		evidencePath := filepath.Join(root, filepath.FromSlash(evidence[0]))

		data, err := os.ReadFile(evidencePath)
		if err != nil {
			t.Errorf("coverage evidence %q is missing: %v", evidence[0], err)
			continue
		}

		text := string(data)

		expectedDeclaration := "func " + evidence[1] + "("
		if strings.HasSuffix(evidence[0], ".lua") {
			expectedDeclaration = `test.case("` + evidence[1] + `"`
		}

		if !strings.Contains(text, expectedDeclaration) {
			t.Errorf("coverage evidence %q does not declare case or test %q", evidence[0], evidence[1])
		}
		// Lua claims and discovered cases are validated from the executed suite
		// metadata by TestLuaSuites. This inventory check only handles file/status
		// validity and Go test evidence, which cannot be loaded as Lua metadata.
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

		if err := scanner.Err(); err != nil {
			_ = file.Close()

			return err
		}

		return file.Close()
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
				"cases = ", "test.case(",
			} {
				if !strings.Contains(text, required) {
					t.Errorf("%s bypasses the versioned Lua test API; missing %q", relative, required)
				}
			}

			for _, forbidden := range []string{
				"initial_data_json", "records_json", "payload = [[", "tests = ", "run = function", " assert(",
				"package.loaded",
				"\n            {\n                submit =", "\n            {\n                submit_system =",
			} {
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

	for _, relativeRoot := range []string{"internal/mod/d2legacy", "internal/acceptance"} {
		goRoot := filepath.Join(root, filepath.FromSlash(relativeRoot))

		err = filepath.WalkDir(goRoot, func(path string, entry os.DirEntry, err error) error {
			if err != nil || entry.IsDir() || !strings.HasSuffix(path, "_test.go") {
				return err
			}

			if path == filepath.Join(root, "internal", "acceptance", "gameplay_ownership_test.go") {
				return nil
			}
			// These acceptance tests intentionally build tiny non-d2legacy modules
			// to verify generic host composition and reload behavior.
			if path == filepath.Join(root, "internal", "acceptance", "alternate_mod_test.go") ||
				path == filepath.Join(root, "internal", "acceptance", "runtime_management_test.go") {
				return nil
			}

			fileSet := token.NewFileSet()

			file, err := parser.ParseFile(fileSet, path, nil, 0)
			if err != nil {
				return err
			}

			ast.Inspect(file, func(node ast.Node) bool {
				literal, ok := node.(*ast.BasicLit)
				if !ok || literal.Kind != token.STRING {
					return true
				}

				value, err := strconv.Unquote(literal.Value)
				if err != nil || !looksLikeEmbeddedLua(value) {
					return true
				}

				relative, _ := filepath.Rel(root, path)
				position := fileSet.Position(literal.Pos())
				t.Errorf(
					"%s:%d embeds Lua source; put the scenario in a checked-in Lua fixture",
					filepath.ToSlash(relative),
					position.Line,
				)

				return true
			})

			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
}

// looksLikeEmbeddedLua uses several syntax signals to avoid flagging ordinary prose that happens to
// contain one Lua keyword. Embedded scenarios belong in checked-in Lua fixtures.
func looksLikeEmbeddedLua(value string) bool {
	signals := 0

	for _, marker := range []string{"local ", "function", "require(", "return ", "end", "~=", " then"} {
		if strings.Contains(value, marker) {
			signals++
		}
	}

	return signals >= 2
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
			// The standalone authoring product owns one explicit VFS export. It is
			// a content path, not a Lua/API namespace, and must remain segment-
			// scoped beneath the editor package.
			text = strings.ReplaceAll(text, "darkmagic/ds1-editor/", "")
			retiredShort := "d" + "m."
			retiredModShort := "d" + "2."
			retiredLong := "dark" + "magic"
			// Only quoted d2.* identifiers are namespaces. Real legacy archive
			// names such as patch_d2.mpq are content filenames, not Lua APIs.
			usesRetiredModShort := strings.Contains(text, `"`+retiredModShort) ||
				strings.Contains(text, `'`+retiredModShort)

			usesRetiredLong := strings.Contains(text, retiredLong+".") ||
				strings.Contains(text, retiredLong+"/")
			if strings.Contains(text, retiredShort) || usesRetiredModShort || usesRetiredLong {
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
	clientRoot := filepath.Join(root, "internal", "app", "clientapp")
	forbidden := []string{
		"RegisterPlayerBasicAttack",
		"RegisterAI",
		"RegisterDeath",
		"game/loot\"",
		"gamemonster.Register,",
		"gameplayer.Register,",
		"gametransition.Register(",
		"gamestate.Register(",
		"gameplayer.RegisterEquipmentProfile(",
	}

	err := filepath.WalkDir(clientRoot, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil || entry.IsDir() {
			return walkErr
		}

		if filepath.Ext(path) != ".go" || strings.HasSuffix(path, "_test.go") {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		text := string(data)

		relative, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}

		for _, marker := range forbidden {
			if strings.Contains(text, marker) {
				t.Errorf("%s reinstalls migrated d2legacy policy through %q", relative, marker)
			}
		}

		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

// Cross-zone policy used to be split between a native proximity command source
// and Lua command application. The fixed-tick d2legacy system now owns the
// complete decision, so no production Go file may recreate those authority
// types or the retired command identity.
func TestWorldTransitionPolicyStaysInD2LegacyLua(t *testing.T) {
	root := repositoryRoot(t)
	for _, relativeRoot := range []string{"internal/app", "internal/mod/d2legacy"} {
		searchRoot := filepath.Join(root, filepath.FromSlash(relativeRoot))

		err := filepath.WalkDir(searchRoot, func(path string, entry os.DirEntry, err error) error {
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

// TestD2MapTilePolicyStaysInD2LegacyLua prevents generic Go world code from deciding Diablo tile
// variants or identities after that policy moved into the first-party mod.
func TestD2MapTilePolicyStaysInD2LegacyLua(t *testing.T) {
	root := repositoryRoot(t)
	for _, relativeRoot := range []string{"internal/game", "internal/mod/d2legacy/adapter"} {
		searchRoot := filepath.Join(root, filepath.FromSlash(relativeRoot))

		err := filepath.WalkDir(searchRoot, func(path string, entry os.DirEntry, err error) error {
			if err != nil || entry.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
				return err
			}

			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}

			forbiddenValues := []string{
				"actOneDirtPathSequence",
				"RealizeActOneDirtPath",
				"ActOneTownEntry",
				"ObjectTypeStatic && object.ID == 2",
			}
			for _, forbidden := range forbiddenValues {
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
					t.Errorf(
						"%s imports %s (%s); generic mechanisms may not gain D2 policy dependencies",
						importer,
						dependency,
						dependencyClass,
					)
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
		searchRoot := filepath.Join(root, filepath.FromSlash(relativeRoot))

		err := filepath.WalkDir(searchRoot, func(path string, entry os.DirEntry, err error) error {
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
	searchRoots := []string{
		"internal/game/ecs",
		"internal/game/session",
		"internal/game/simulation",
		"internal/game/worldgen",
		"internal/runtime/lua",
	}

	for _, relativeRoot := range searchRoots {
		searchRoot := filepath.Join(root, filepath.FromSlash(relativeRoot))

		err := filepath.WalkDir(searchRoot, func(path string, entry os.DirEntry, err error) error {
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
					t.Errorf(
						"%s names Diablo policy %q; generic engine code must use mod-neutral vocabulary",
						filepath.ToSlash(relative),
						token,
					)
				}
			}

			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
}

// readOwnershipRules parses the checked-in tab-separated inventory and rejects malformed rows early,
// so later architectural failures refer to trusted rule data.
func readOwnershipRules(t *testing.T, path string) []ownershipRule {
	t.Helper()

	file, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		if err := file.Close(); err != nil {
			t.Errorf("close ownership inventory: %v", err)
		}
	}()

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

// readDependencyDebt loads the exact grandfathered mechanism-to-policy imports. The set is a ratchet,
// not an allowlist for newly introduced violations.
func readDependencyDebt(t *testing.T, path string) map[string]struct{} {
	t.Helper()

	file, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		if err := file.Close(); err != nil {
			t.Errorf("close dependency-debt inventory: %v", err)
		}
	}()

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

// ownershipClassForPath returns the inventory class for one repository-relative path; callers have
// already enforced that production files match exactly one rule.
func ownershipClassForPath(rules []ownershipRule, path string) string {
	for _, rule := range rules {
		if strings.HasPrefix(path, rule.prefix) {
			return rule.class
		}
	}

	return ""
}

// isProductionOwnershipFile excludes tests, documentation, and fixtures because the ownership
// inventory governs executable Go and Lua behavior only.
func isProductionOwnershipFile(path string) bool {
	if strings.HasSuffix(path, "_test.go") || strings.HasSuffix(path, "_test.lua") {
		return false
	}

	if strings.Contains(filepath.ToSlash(path), "/lua/d2legacy/tests/") {
		return false
	}

	return strings.HasSuffix(path, ".go") || strings.HasSuffix(path, ".lua")
}
