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

// TestLuaNamespacesDescribeOwnership prevents the retired two-letter
// abbreviation from blurring generic engine APIs with the d2legacy mod again. Engine doors
// use engine.*; Diablo runtime state and bundled modules use d2legacy.* or d2legacy.*.
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
			if strings.Contains(text, retiredShort) || strings.Contains(text, retiredModShort) || strings.Contains(text, retiredLong+".") || strings.Contains(text, retiredLong+"/") {
				relative, _ := filepath.Rel(root, path)
				t.Errorf("%s uses a retired Lua namespace; use engine.*, d2legacy.*, or d2legacy.*", filepath.ToSlash(relative))
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
	if strings.HasSuffix(path, "_test.go") {
		return false
	}
	return strings.HasSuffix(path, ".go") || strings.HasSuffix(path, ".lua")
}
