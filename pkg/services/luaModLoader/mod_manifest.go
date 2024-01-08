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
	Requires []string
}

func (m *Manifest) ID() string {
	return fmt.Sprintf("%s (%s)", m.Name, m.Version)
}

func (m *Manifest) String() string {
	return m.ID()
}

func (m *Manifest) ApiKey() string {
	const (
		regexBadCharacters   = "[^a-zA-Z0-9]"
		regexApiKeyValidator = "^[a-zA-Z][a-zA-Z0-9]+$"
	)

	replacer := regexp.MustCompile(regexBadCharacters)
	validator := regexp.MustCompile(regexApiKeyValidator)

	apiKey := strings.ToLower(m.ID())
	apiKey = strings.ReplaceAll(apiKey, " ", "_")

	apiKey = string(replacer.ReplaceAll([]byte(apiKey), []byte("")))
	if !validator.Match([]byte(apiKey)) {
		panic("manifest has invalid api key")
	}

	return apiKey
}
