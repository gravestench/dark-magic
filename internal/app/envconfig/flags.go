package envconfig

import (
	"errors"
	"strings"
)

// ExplicitPath extracts --env-file without consuming the command's other flags.
func ExplicitPath(arguments []string) (string, error) {
	var selectedPath string
	for index := 0; index < len(arguments); index++ {
		argument := arguments[index]
		if argument == "--" {
			break
		}
		value, consumed, err := environmentFlagValue(arguments, index)
		if err != nil {
			return "", err
		}
		if !consumed {
			continue
		}
		if argument == "--env-file" || argument == "-env-file" {
			index++
		}
		if selectedPath != "" {
			return "", errors.New("--env-file may only be supplied once")
		}
		selectedPath = value
	}
	return selectedPath, nil
}

// environmentFlagValue recognizes one supported --env-file spelling.
func environmentFlagValue(arguments []string, index int) (string, bool, error) {
	argument := arguments[index]
	var value string
	switch {
	case argument == "--env-file" || argument == "-env-file":
		if index+1 >= len(arguments) {
			return "", true, errors.New("--env-file requires a path")
		}
		value = arguments[index+1]
	case strings.HasPrefix(argument, "--env-file="):
		value = strings.TrimPrefix(argument, "--env-file=")
	case strings.HasPrefix(argument, "-env-file="):
		value = strings.TrimPrefix(argument, "-env-file=")
	default:
		return "", false, nil
	}
	if strings.TrimSpace(value) == "" {
		return "", true, errors.New("--env-file requires a non-empty path")
	}
	return value, true, nil
}
