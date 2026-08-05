package luaModLoader

import (
	"fmt"
	"regexp"
	"strings"
)

type Manifest struct {
	Name     string
	Version  string
	Sources  map[string][]string // group names -> path/url
	rootDir  string              // assigned at runtime
	Enabled  bool
	Requires []string // lua globals that must exist before init is called
}

func (m *Manifest) Validate() error {
	if strings.TrimSpace(m.Name) == "" {
		return fmt.Errorf("manifest name is required")
	}
	if strings.TrimSpace(m.Version) == "" {
		return fmt.Errorf("manifest version is required")
	}
	if !regexp.MustCompile(`^[a-z][a-z0-9_]+$`).MatchString(m.apiKey()) {
		return fmt.Errorf("manifest name and version do not produce a valid API key")
	}
	for _, requirement := range m.Requires {
		if strings.TrimSpace(requirement) == "" {
			return fmt.Errorf("manifest contains an empty requirement")
		}
	}
	return nil
}

func (m *Manifest) ID() string {
	return fmt.Sprintf("%s (%s)", m.Name, m.Version)
}

func (m *Manifest) String() string {
	return m.ID()
}

func (m *Manifest) ApiKey() string {
	apiKey := m.apiKey()
	if err := m.Validate(); err != nil {
		panic(err)
	}
	return apiKey
}

func (m *Manifest) apiKey() string {
	const regexBadCharacters = "[^a-zA-Z0-9]"

	replacer := regexp.MustCompile(regexBadCharacters)

	apiKey := strings.ToLower(m.ID())
	apiKey = strings.ReplaceAll(apiKey, " ", "_")

	apiKey = string(replacer.ReplaceAll([]byte(apiKey), []byte("")))

	return apiKey
}
