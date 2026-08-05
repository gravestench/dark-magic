package fileLoader

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/gravestench/servicemesh"

	"github.com/gravestench/dark-magic/pkg/services/common"
)

const (
	DefaultGroup            = ""
	EventServiceInitialized = "file loader initialized"
)

var _ CompositeFilesystemLoader = &Service{}

type Service struct {
	common.Service
	fsGroups map[string][]Source
	mux      sync.Mutex
}

func New() *Service {
	return &Service{fsGroups: make(map[string][]Source)}
}

func (s *Service) GetSources(group string) (sources []fs.FS) {
	for _, source := range s.fsGroups[group] {
		f, err := source.Filesystem()
		if err != nil {
			s.Logger().Error("getting source filesystem", "group", group, "source", source.Path)
			continue
		}
		sources = append(sources, f)
	}

	return
}

func (s *Service) Load(path string, groups ...string) (fs.File, error) {
	return s.FromGroups(groups...).Open(path)
}

func (s *Service) Init(mesh servicemesh.Mesh) {
	s.mux.Lock()
	s.fsGroups = make(map[string][]Source)
	s.mux.Unlock()

	if err := s.AddSource(NewSource(".")); err != nil {
		s.Logger().Error("adding source", "error", err)
		mesh.Shutdown()
		return
	}
	if err := s.addMPQDirectorySources(os.Getenv("MPQ_DIRECTORY")); err != nil {
		s.Logger().Warn("loading MPQ directory", "error", err)
	}

	mesh.Events().Emit(EventServiceInitialized)
}

func (s *Service) addMPQDirectorySources(directory string) error {
	if directory == "" {
		return nil
	}
	info, err := os.Stat(directory)
	if err != nil {
		return fmt.Errorf("checking MPQ directory: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("MPQ_DIRECTORY %q is not a directory", directory)
	}
	// Earlier sources win. Expansion patches override expansion data, which
	// override classic patches and base archives.
	priority := []string{
		"patch_d2.mpq", "d2exp.mpq", "d2data.mpq", "d2char.mpq",
		"d2music.mpq", "d2sfx.mpq", "d2speech.mpq", "d2video.mpq",
		"d2xmusic.mpq", "d2xtalk.mpq", "d2xvideo.mpq",
	}
	for _, name := range priority {
		path := filepath.Join(directory, name)
		if _, err := os.Stat(path); err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return fmt.Errorf("checking MPQ %q: %w", path, err)
		}
		if err := s.AddSource(NewSource(path)); err != nil {
			return fmt.Errorf("adding MPQ %q: %w", path, err)
		}
	}
	return nil
}

func (s *Service) Name() string {
	return "Composite FS Loader"
}

func (s *Service) Ready() bool {
	return true
}

func (s *Service) Groups() (groups []string) {
	s.mux.Lock()
	defer s.mux.Unlock()

	for group := range s.fsGroups {
		groups = append(groups, group)
	}

	sort.Strings(groups)

	return
}

func (s *Service) AddSource(src Source) error {
	return s.AddSourceToGroup(src, DefaultGroup)
}

func (s *Service) AddSourceToGroup(src Source, group string) error {
	_ = s.RemoveSourceFromGroup(src, group)

	s.mux.Lock()
	defer s.mux.Unlock()
	if s.fsGroups == nil {
		s.fsGroups = make(map[string][]Source)
	}

	s.fsGroups[group] = append(s.fsGroups[group], src)

	return nil
}

func (s *Service) RemoveSource(src Source) error {
	return s.RemoveSourceFromGroup(src, DefaultGroup)
}

func (s *Service) RemoveSourceFromGroup(src Source, group string) error {
	idx := s.srcIndexFromGroup(src, group)
	if idx < 0 {
		return fmt.Errorf("not found")
	}

	s.mux.Lock()
	defer s.mux.Unlock()

	s.fsGroups[group] = append(s.fsGroups[group][:idx], s.fsGroups[group][idx+1:]...)

	return nil
}

func (s *Service) srcIndexFromGroup(src Source, group string) int {
	s.mux.Lock()
	defer s.mux.Unlock()

	for idx, existing := range s.fsGroups[group] {
		if src.Path == existing.Path {
			return idx
		}
	}

	return -1
}

func (s *Service) RemoveGroup(group string) error {
	s.mux.Lock()
	defer s.mux.Unlock()

	if _, found := s.fsGroups[group]; !found {
		return fmt.Errorf("group does not exist")
	}

	delete(s.fsGroups, group)

	return nil
}

func (s *Service) Open(path string, groups ...string) (fs.File, error) {
	if len(groups) < 1 {
		groups = s.Groups()
	}

	s.mux.Lock()
	groupSources := make(map[string][]Source, len(groups))
	for _, group := range groups {
		groupSources[group] = append([]Source(nil), s.fsGroups[group]...)
	}
	s.mux.Unlock()

	for _, groupKey := range groups {
		if group, found := groupSources[groupKey]; found {
			for _, src := range group {
				srcFS, err := src.Filesystem()
				if err != nil {
					continue
				}

				f, err := srcFS.Open(path)
				if err != nil {
					continue
				}

				return f, nil
			}
		}
	}

	return nil, fmt.Errorf("file not found")
}

type closure struct {
	s      *Service
	groups []string
}

func (c closure) Open(name string) (fs.File, error) {
	return c.s.Open(name, c.groups...) // default, try to open file from anywhere
}

func (s *Service) FromGroup(group string) fs.FS {
	return &closure{
		s:      s,
		groups: []string{group},
	}
}

func (s *Service) FromGroups(groups ...string) fs.FS {
	cs := make(closures, 0)

	if len(groups) < 1 {
		groups = []string{""}
	}

	for _, group := range groups {
		cs = append(cs, closure{
			s:      s,
			groups: []string{group},
		})
	}

	return cs
}

type closures []closure

func (cs closures) Open(name string) (fs.File, error) {
	lookup := make(map[string]bool)
	groups := make([]string, 0)

	for _, c := range cs {
		for _, group := range c.groups {
			if _, found := lookup[group]; found {
				continue
			}
			lookup[group] = true
		}

		if f, err := c.Open(name); err == nil {
			return f, err
		}
	}

	return nil, fmt.Errorf("file %q not found in groups (%s)", name, strings.Join(groups, ", "))
}
