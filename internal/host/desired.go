package host

import (
	"fmt"
	"strings"
)

// ParseDesired parses a comma-separated component list. An empty value enables
// defaults; "none" explicitly disables every optional component.
func ParseDesired(value string, defaults ...string) (map[string]bool, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		value = strings.Join(defaults, ",")
	}
	result := make(map[string]bool)
	if strings.EqualFold(value, "none") {
		return result, nil
	}
	for _, raw := range strings.Split(value, ",") {
		id := strings.TrimSpace(raw)
		if id == "" {
			return nil, fmt.Errorf("host: empty component ID in desired-state configuration")
		}
		result[id] = true
	}
	return result, nil
}
