// Package branding provides artwork embedded in Dark Magic binaries.
package branding

import _ "embed"

//go:embed window-icon.png
var windowIconPNG []byte

// WindowIconPNG returns the PNG-encoded application window icon.
func WindowIconPNG() []byte {
	return windowIconPNG
}
