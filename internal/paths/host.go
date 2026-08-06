package paths

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode"
)

// ExpandHost expands shell-style aliases in a host filesystem path on every
// platform. MPQ-internal paths must not pass through this function.
//
// Supported forms are ~, ~/path, ~\path, $NAME, ${NAME}, and %NAME%. Missing
// environment variables are errors so a typo cannot silently redirect I/O.
func ExpandHost(name string) (string, error) {
	expanded, err := expandPercentEnvironment(name)
	if err != nil {
		return "", err
	}
	expanded, err = expandDollarEnvironment(expanded)
	if err != nil {
		return "", err
	}
	if expanded != name {
		expanded = filepath.FromSlash(strings.ReplaceAll(expanded, `\`, "/"))
	}
	if expanded == "" || expanded[0] != '~' {
		return expanded, nil
	}
	if expanded != "~" && !strings.HasPrefix(expanded, "~/") && !strings.HasPrefix(expanded, `~\`) {
		return "", fmt.Errorf("paths: named home directories are unsupported in %q", name)
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("paths: resolve home directory: %w", err)
	}
	if expanded == "~" {
		return home, nil
	}
	parts := strings.FieldsFunc(expanded[2:], func(r rune) bool { return r == '/' || r == '\\' })
	return filepath.Join(append([]string{home}, parts...)...), nil
}

func expandDollarEnvironment(name string) (string, error) {
	var missing string
	expanded := os.Expand(name, func(key string) string {
		value, exists := os.LookupEnv(key)
		if !exists && missing == "" {
			missing = key
		}
		return value
	})
	if missing != "" {
		return "", fmt.Errorf("paths: environment variable %q is not set", missing)
	}
	return expanded, nil
}

func expandPercentEnvironment(name string) (string, error) {
	var result strings.Builder
	for index := 0; index < len(name); {
		if name[index] != '%' {
			result.WriteByte(name[index])
			index++
			continue
		}
		end := strings.IndexByte(name[index+1:], '%')
		if end < 0 {
			result.WriteByte(name[index])
			index++
			continue
		}
		end += index + 1
		key := name[index+1 : end]
		if key == "" || strings.IndexFunc(key, func(r rune) bool { return r != '_' && !unicode.IsLetter(r) && !unicode.IsDigit(r) }) >= 0 {
			result.WriteString(name[index : end+1])
			index = end + 1
			continue
		}
		value, exists := os.LookupEnv(key)
		if !exists {
			return "", fmt.Errorf("paths: environment variable %q is not set", key)
		}
		result.WriteString(value)
		index = end + 1
	}
	return result.String(), nil
}
