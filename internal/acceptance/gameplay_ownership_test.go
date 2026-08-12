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
		"internal/content/shim/lua",
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
