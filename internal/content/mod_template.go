package content

import (
	"embed"
	"io/fs"
)

//go:embed mod_template
var embeddedModTemplate embed.FS

// ModTemplate returns the bundled empty mod reference tree. It is not selected
// by the application; callers use it as validated starter content for tooling,
// examples, and new mod composition.
func ModTemplate() fs.FS {
	mod, err := fs.Sub(embeddedModTemplate, "mod_template")
	if err != nil {
		panic(err)
	}
	return mod
}
