package fileLoader

import (
	"io/fs"
)

type Dependency = CompositeFilesystemLoader

type CompositeFilesystemLoader interface {
	FromGroup(group string) fs.FS
	FromGroups(group ...string) fs.FS
	Groups() []string
	GetSources(group string) []fs.FS
	AddSource(src Source) error
	AddSourceToGroup(src Source, group string) error
	RemoveSource(src Source) error
	RemoveSourceFromGroup(src Source, group string) error
	RemoveGroup(group string) error
	Load(path string, groups ...string) (fs.File, error)
}
