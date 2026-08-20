// Package branding provides artwork embedded in Dark Magic binaries.
package branding

import _ "embed"

//go:embed window-icon.png
var windowIconPNG []byte

// WindowIconPNG returns the embedded application window icon in PNG format.
// The returned slice aliases package storage, so callers must treat it as read-only to keep future reads stable.
func WindowIconPNG() []byte {
	return windowIconPNG
}
