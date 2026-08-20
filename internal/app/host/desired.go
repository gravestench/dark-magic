package host

import (
	"fmt"
	"strings"
)

// ParseDesired distinguishes omitted configuration from the explicit "none" sentinel and rejects empty list slots.
func ParseDesired(value string, defaults ...string) (map[string]bool, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		value = strings.Join(defaults, ",")
	}

	desired := make(map[string]bool)
	if strings.EqualFold(value, "none") {
		return desired, nil
	}

	for _, raw := range strings.Split(value, ",") {
		id := strings.TrimSpace(raw)
		if id == "" {
			return nil, fmt.Errorf("host: empty component ID in desired-state configuration")
		}

		desired[id] = true
	}

	return desired, nil
}
