package content

import (
	"embed"
	"io/fs"
)

//go:embed shim
var embeddedShim embed.FS

// Shim returns the redistributable first-party Dark Magic content tree.
func Shim() fs.FS {
	shim, err := fs.Sub(embeddedShim, "shim")
	if err != nil {
		panic(err)
	}
	return shim
}
