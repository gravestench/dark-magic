package clientapp

import (
	"crypto/sha256"
	"encoding/hex"
	"io/fs"
	"sort"
	"strings"

	"github.com/gravestench/dark-magic/internal/game/simulation"
)

// d2legacyIdentity pins the exact authoritative Lua bytes used by a session.
// File order is sorted so the hash is independent of filesystem traversal.
func (app *application) d2legacyIdentity() (simulation.RuntimeIdentity, error) {
	var names []string
	err := fs.WalkDir(app.options.Content, "lua/d2legacy", func(name string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !entry.IsDir() && strings.HasSuffix(name, ".lua") {
			names = append(names, name)
		}
		return nil
	})
	if err != nil {
		return simulation.RuntimeIdentity{}, err
	}
	sort.Strings(names)
	hash := sha256.New()
	for _, name := range names {
		data, err := fs.ReadFile(app.options.Content, name)
		if err != nil {
			return simulation.RuntimeIdentity{}, err
		}
		hash.Write([]byte(name))
		hash.Write([]byte{0})
		hash.Write(data)
	}
	digest := hex.EncodeToString(hash.Sum(nil))
	return simulation.RuntimeIdentity{
		ModID: "d2legacy", ContractVersion: "v1",
		PackageHash: digest, AuthoritativeHash: digest,
		ConfigurationHash: "d2legacy/default/v1",
		CapabilityVersions: map[string]string{
			"dm.authority_command": "v1", "dm.authority_random": "v1",
			"dm.authority_state": "v1", "dm.ecs": "v1", "dm.records": "v1",
		},
	}, nil
}
