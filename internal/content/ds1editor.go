package content

import (
	"embed"
	"io/fs"
)

//go:embed ds1editor
var embeddedDS1Editor embed.FS

// DS1Editor returns the standalone authoring package. Its Lua and generated
// art are owned independently from d2legacy's game presentation package.
func DS1Editor() fs.FS {
	content, err := fs.Sub(embeddedDS1Editor, "ds1editor")
	if err != nil {
		panic(err)
	}
	return content
}
