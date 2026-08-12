package content

import (
	"embed"
	"io/fs"
)

//go:embed shim
var embeddedShim embed.FS

// D2Legacy returns the redistributable first-party Diablo II legacy mod tree.
// The engine embeds these development assets for now; their ownership is still
// the d2legacy mod's, not the generic host's.
func D2Legacy() fs.FS {
	mod, err := fs.Sub(embeddedShim, "shim")
	if err != nil {
		panic(err)
	}
	return mod
}
