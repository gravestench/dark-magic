// Package envconfig installs and loads the per-process environment files used
// by Dark Magic composition roots. Exported process variables remain
// authoritative; files provide local defaults without replacing CLI flags.
package envconfig

import (
	"bufio"
	"embed"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"

	darkpaths "github.com/gravestench/dark-magic/internal/paths"
)

//go:embed templates/*.env
var templates embed.FS

var roles = map[string]struct{}{"client": {}, "server": {}, "realm": {}}

type Result struct {
	Role        string
	DefaultPath string
	LoadedPath  string
	Created     bool
}

// Duration returns a positive duration from the process environment or the
// supplied positive fallback when the variable is empty.
func Duration(name string, fallback time.Duration) (time.Duration, error) {
	if strings.TrimSpace(name) == "" || fallback <= 0 {
		return 0, errors.New("environment duration requires a name and positive fallback")
	}
	value := strings.TrimSpace(os.Getenv(name))
	if value == "" {
		return fallback, nil
	}
	parsed, err := time.ParseDuration(value)
	if err != nil || parsed <= 0 {
		return 0, fmt.Errorf("invalid %s duration %q", name, value)
	}
	return parsed, nil
}

// Bootstrap installs the default template, selects an optional --env-file,
// and loads values which are not already present in the process environment.
func Bootstrap(role string, arguments []string) (Result, error) {
	defaultPath, created, err := Install(role)
	if err != nil {
		return Result{}, err
	}
	explicit, err := ExplicitPath(arguments)
	if err != nil {
		return Result{}, err
	}
	loadedPath := defaultPath
	if explicit != "" {
		loadedPath, err = darkpaths.ExpandHost(explicit)
		if err != nil {
			return Result{}, fmt.Errorf("environment file: %w", err)
		}
	}
	if err := Load(loadedPath); err != nil {
		return Result{}, err
	}
	return Result{Role: role, DefaultPath: defaultPath, LoadedPath: loadedPath, Created: created}, nil
}

// ExplicitPath extracts --env-file without consuming ordinary flags. The
// command's flag set parses it normally after the environment has been loaded.
func ExplicitPath(arguments []string) (string, error) {
	var result string
	for index := 0; index < len(arguments); index++ {
		argument := arguments[index]
		if argument == "--" {
			break
		}
		var value string
		switch {
		case argument == "--env-file" || argument == "-env-file":
			index++
			if index >= len(arguments) {
				return "", errors.New("--env-file requires a path")
			}
			value = arguments[index]
		case strings.HasPrefix(argument, "--env-file="):
			value = strings.TrimPrefix(argument, "--env-file=")
		case strings.HasPrefix(argument, "-env-file="):
			value = strings.TrimPrefix(argument, "-env-file=")
		default:
			continue
		}
		if strings.TrimSpace(value) == "" {
			return "", errors.New("--env-file requires a non-empty path")
		}
		if result != "" {
			return "", errors.New("--env-file may only be supplied once")
		}
		result = value
	}
	return result, nil
}

func Install(role string) (string, bool, error) {
	if _, found := roles[role]; !found {
		return "", false, fmt.Errorf("unknown environment role %q", role)
	}
	directory, err := configDirectory()
	if err != nil {
		return "", false, err
	}
	if err := os.MkdirAll(directory, 0o700); err != nil {
		return "", false, fmt.Errorf("create environment directory: %w", err)
	}
	if err := os.Chmod(directory, 0o700); err != nil {
		return "", false, fmt.Errorf("secure environment directory: %w", err)
	}
	path := filepath.Join(directory, role+".env")
	if _, err := os.Stat(path); err == nil {
		if err := os.Chmod(path, 0o600); err != nil {
			return "", false, fmt.Errorf("secure environment file %q: %w", path, err)
		}
		return path, false, nil
	} else if !errors.Is(err, fs.ErrNotExist) {
		return "", false, fmt.Errorf("inspect environment file %q: %w", path, err)
	}
	data, err := templates.ReadFile("templates/" + role + ".env")
	if err != nil {
		return "", false, err
	}
	if err := writePrivate(path, data); err != nil {
		return "", false, fmt.Errorf("install environment file %q: %w", path, err)
	}
	return path, true, nil
}

// Update installs a role file when necessary and replaces only known template
// keys, preserving its comments and unrelated formatting.
func Update(role string, updates map[string]string) (string, error) {
	path, _, err := Install(role)
	if err != nil {
		return "", err
	}
	templateData, err := templates.ReadFile("templates/" + role + ".env")
	if err != nil {
		return "", err
	}
	allowed, err := Parse(strings.NewReader(string(templateData)))
	if err != nil {
		return "", err
	}
	for key := range updates {
		if _, found := allowed[key]; !found {
			return "", fmt.Errorf("environment variable %q is not part of the %s template", key, role)
		}
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	remaining := make(map[string]string, len(updates))
	for key, value := range updates {
		remaining[key] = value
	}
	lines := strings.Split(strings.TrimSuffix(string(data), "\n"), "\n")
	for index, line := range lines {
		trimmed := strings.TrimSpace(line)
		separator := strings.IndexByte(trimmed, '=')
		if separator <= 0 {
			continue
		}
		key := strings.TrimSpace(trimmed[:separator])
		value, found := remaining[key]
		if !found {
			continue
		}
		lines[index] = key + "=" + strconv.Quote(value)
		delete(remaining, key)
	}
	keys := make([]string, 0, len(remaining))
	for key := range remaining {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		lines = append(lines, key+"="+strconv.Quote(remaining[key]))
	}
	return path, writePrivate(path, []byte(strings.Join(lines, "\n")+"\n"))
}

func configDirectory() (string, error) {
	if configured := strings.TrimSpace(os.Getenv("DARK_MAGIC_CONFIG_DIR")); configured != "" {
		expanded, err := darkpaths.ExpandHost(configured)
		if err != nil {
			return "", fmt.Errorf("environment directory: %w", err)
		}
		return expanded, nil
	}
	root, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("resolve environment directory: %w", err)
	}
	return filepath.Join(root, "dark-magic"), nil
}

func Load(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open environment file %q: %w", path, err)
	}
	defer file.Close()
	values, err := Parse(file)
	if err != nil {
		return fmt.Errorf("parse environment file %q: %w", path, err)
	}
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		if _, exists := os.LookupEnv(key); exists {
			continue
		}
		if err := os.Setenv(key, values[key]); err != nil {
			return fmt.Errorf("set environment variable %q: %w", key, err)
		}
	}
	return nil
}

func Parse(reader io.Reader) (map[string]string, error) {
	values := make(map[string]string)
	scanner := bufio.NewScanner(reader)
	scanner.Buffer(make([]byte, 4096), 256<<10)
	for lineNumber := 1; scanner.Scan(); lineNumber++ {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "export ") {
			line = strings.TrimSpace(strings.TrimPrefix(line, "export "))
		}
		separator := strings.IndexByte(line, '=')
		if separator <= 0 {
			return nil, fmt.Errorf("line %d: expected NAME=VALUE", lineNumber)
		}
		key := strings.TrimSpace(line[:separator])
		if !validKey(key) {
			return nil, fmt.Errorf("line %d: invalid variable name %q", lineNumber, key)
		}
		value, err := parseValue(strings.TrimSpace(line[separator+1:]))
		if err != nil {
			return nil, fmt.Errorf("line %d: %w", lineNumber, err)
		}
		values[key] = value
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return values, nil
}

func validKey(value string) bool {
	for index, character := range value {
		if character != '_' && !unicode.IsLetter(character) && (index == 0 || !unicode.IsDigit(character)) {
			return false
		}
	}
	return value != ""
}

func parseValue(value string) (string, error) {
	if value == "" {
		return "", nil
	}
	if value[0] == '\'' {
		if len(value) < 2 || value[len(value)-1] != '\'' {
			return "", errors.New("unterminated single-quoted value")
		}
		return value[1 : len(value)-1], nil
	}
	if value[0] == '"' {
		if len(value) < 2 || value[len(value)-1] != '"' {
			return "", errors.New("unterminated double-quoted value")
		}
		decoded, err := strconv.Unquote(value)
		if err != nil {
			return "", fmt.Errorf("invalid double-quoted value: %w", err)
		}
		return decoded, nil
	}
	if index := strings.Index(value, " #"); index >= 0 {
		value = strings.TrimSpace(value[:index])
	}
	return value, nil
}

func writePrivate(path string, data []byte) error {
	directory := filepath.Dir(path)
	temporary, err := os.CreateTemp(directory, ".env-*")
	if err != nil {
		return err
	}
	temporaryPath := temporary.Name()
	defer os.Remove(temporaryPath)
	if err := temporary.Chmod(0o600); err == nil {
		_, err = temporary.Write(data)
	}
	if syncErr := temporary.Sync(); err == nil {
		err = syncErr
	}
	if closeErr := temporary.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		return err
	}
	return os.Rename(temporaryPath, path)
}
