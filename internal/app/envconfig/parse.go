package envconfig

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"unicode"
)

// Parse reads the supported human-readable dotenv syntax from a stream.
func Parse(reader io.Reader) (map[string]string, error) {
	values := make(map[string]string)
	scanner := bufio.NewScanner(reader)
	scanner.Buffer(make([]byte, 4096), 256<<10)
	for lineNumber := 1; scanner.Scan(); lineNumber++ {
		key, value, present, err := parseLine(scanner.Text())
		if err != nil {
			return nil, fmt.Errorf("line %d: %w", lineNumber, err)
		}
		if present {
			values[key] = value
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return values, nil
}

// parseLine parses one assignment and identifies comments or blank lines.
func parseLine(line string) (string, string, bool, error) {
	line = strings.TrimSpace(line)
	if line == "" || strings.HasPrefix(line, "#") {
		return "", "", false, nil
	}
	line = strings.TrimSpace(strings.TrimPrefix(line, "export "))
	separator := strings.IndexByte(line, '=')
	if separator <= 0 {
		return "", "", false, errors.New("expected NAME=VALUE")
	}
	key := strings.TrimSpace(line[:separator])
	if !validKey(key) {
		return "", "", false, fmt.Errorf("invalid variable name %q", key)
	}
	value, err := parseValue(strings.TrimSpace(line[separator+1:]))
	if err != nil {
		return "", "", false, err
	}
	return key, value, true, nil
}

// validKey reports whether a name follows shell-compatible environment syntax.
func validKey(value string) bool {
	for index, character := range value {
		validCharacter := character == '_' || unicode.IsLetter(character)
		if index > 0 {
			validCharacter = validCharacter || unicode.IsDigit(character)
		}
		if !validCharacter {
			return false
		}
	}
	return value != ""
}

// parseValue decodes quoting and strips supported trailing comments.
func parseValue(value string) (string, error) {
	if value == "" {
		return "", nil
	}
	switch value[0] {
	case '\'':
		return parseSingleQuotedValue(value)
	case '"':
		return parseDoubleQuotedValue(value)
	default:
		return unquotedValue(value), nil
	}
}

// parseSingleQuotedValue removes literal single-quote delimiters.
func parseSingleQuotedValue(value string) (string, error) {
	if len(value) < 2 || value[len(value)-1] != '\'' {
		return "", errors.New("unterminated single-quoted value")
	}
	return value[1 : len(value)-1], nil
}

// parseDoubleQuotedValue delegates escape decoding to Go's string parser.
func parseDoubleQuotedValue(value string) (string, error) {
	if len(value) < 2 || value[len(value)-1] != '"' {
		return "", errors.New("unterminated double-quoted value")
	}
	decoded, err := strconv.Unquote(value)
	if err != nil {
		return "", fmt.Errorf("invalid double-quoted value: %w", err)
	}
	return decoded, nil
}

// unquotedValue removes a trailing comment introduced by whitespace and '#'.
func unquotedValue(value string) string {
	if index := strings.Index(value, " #"); index >= 0 {
		return strings.TrimSpace(value[:index])
	}
	return value
}
