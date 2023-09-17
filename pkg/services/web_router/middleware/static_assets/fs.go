package static_assets

import (
	"embed"
	"fmt"
	"io/fs"
)

//go:embed embedded
var embedded embed.FS

const prefix = "embedded"

func prefixed(s string) string {
	for {
		if len(s) < 2 {
			break
		}

		if string(s[0]) == "/" {
			s = s[1:]
			continue
		}

		break
	}

	return fmt.Sprintf("%s/%s", prefix, s)
}

func (m *Middleware) Open(name string) (fs.File, error) {
	name = prefixed(name)

	f, err := embedded.Open(name)
	if err != nil {
		err = fmt.Errorf("opening file in embedded filesystem: %v", err)
	}

	return f, err
}

func (m *Middleware) ReadDir(name string) ([]fs.DirEntry, error) {
	name = prefixed(name)

	list, err := embedded.ReadDir(name)
	if err != nil {
		err = fmt.Errorf("reading directory in embedded filesystem: %v", err)
	}

	return list, err
}

func (m *Middleware) ReadFile(name string) ([]byte, error) {
	name = prefixed(name)

	data, err := embedded.ReadFile(name)
	if err != nil {
		err = fmt.Errorf("reading file from embedded filesystem: %v", err)
	}

	return data, err
}
