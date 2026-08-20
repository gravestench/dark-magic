package assetdecode

import (
	"fmt"
	"io/fs"

	cof "github.com/gravestench/cof"
)

// COF reads composite ordering, layer, timing, and event metadata without
// exposing archive ownership to the codec.
func COF(source fs.FS, name string) (*cof.COF, error) {
	file, err := source.Open(name)
	if err != nil {
		return nil, fmt.Errorf("cof %q: %w", name, err)
	}
	defer file.Close() //nolint:errcheck // Decode errors retain precedence over read-only close failures.

	asset, err := cof.UnmarshalReader(file)
	if err != nil {
		return nil, fmt.Errorf("cof %q: %w", name, err)
	}

	return asset, nil
}

// AnimationData reads the global typed timing and event catalog directly from
// its stream, so callers need not buffer the binary file or know its block layout.
func AnimationData(source fs.FS, name string) (*cof.AnimationData, error) {
	file, err := source.Open(name)
	if err != nil {
		return nil, fmt.Errorf("animation data %q: %w", name, err)
	}
	defer file.Close() //nolint:errcheck // Decode errors retain precedence over read-only close failures.

	asset, err := cof.LoadReader(file)
	if err != nil {
		return nil, fmt.Errorf("animation data %q: %w", name, err)
	}

	return asset, nil
}
