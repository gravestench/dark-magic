package mpqLoader

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/gravestench/mpq"
	"k8s.io/utils/strings/slices"

	"github.com/gravestench/servicemesh"

	"github.com/gravestench/dark-magic/pkg/services/configFile"
)

type Service struct {
	mesh          servicemesh.Mesh
	logger        *slog.Logger
	configManager configFile.Dependency
	config        *configFile.Config
	archives      map[string]*mpq.MPQ
	ordering      []string
	mux           sync.Mutex
}

func (s *Service) Init(mesh servicemesh.Mesh) {
	s.mesh = mesh
	s.archives = make(map[string]*mpq.MPQ)
	s.ordering = make([]string, 0)

	go s.loadArchivesFromFiles()
}

func (s *Service) Name() string {
	return "MPQ Archive Loader"
}

func (s *Service) Ready() bool {
	if s.config == nil {
		return false
	}

	return true
}

func (s *Service) SetLogger(logger *slog.Logger) {
	s.logger = logger
}

func (s *Service) Logger() *slog.Logger {
	return s.logger
}

func (s *Service) Archives() map[string]*mpq.MPQ {
	return s.archives
}

func (s *Service) AddArchive(filepath string) error {
	s.mux.Lock()
	defer s.mux.Unlock()

	if slices.Contains(s.ordering, filepath) {
		return fmt.Errorf("mpq already loaded for filepath %s", filepath)
	}

	loaded, err := mpq.FromFile(filepath)
	if err != nil {
		return fmt.Errorf("loading archive", "error", err)
	}

	s.archives[filepath] = loaded
	s.ordering = append(s.ordering, filepath)

	return nil
}

func (s *Service) RemoveArchive(filepath string) error {
	s.mux.Lock()
	defer s.mux.Unlock()

	if !slices.Contains(s.ordering, filepath) {
		return fmt.Errorf("mpq not found for filepath %q", filepath)
	}

	delete(s.archives, filepath)
	removeStringFromArray(filepath, s.ordering)

	return nil
}

func (s *Service) Load(filepath string) (io.Reader, error) {
	//s.mux.Lock()
	//defer s.mux.Unlock()

	allErrors := make([]error, 0)

	filepath = strings.ReplaceAll(filepath, "/data", `data`)
	filepath = strings.ReplaceAll(filepath, "/", `\`)

	for _, key := range s.ordering {
		archive := s.archives[key]

		if data, err := archive.ReadFileStream(filepath); err == nil {
			return data, nil
		}

		allErrors = append(allErrors, fmt.Errorf("not found in %v", key))
	}

	// if we get to this point it is because we have not been able to load the
	// file in any of the known archives we are loading from

	var errMsg string

	for _, err := range allErrors {
		if err == nil {
			continue
		}

		if errMsg == "" {
			errMsg = fmt.Sprintf("not found in", "error", err)
			continue
		}

		errMsg = fmt.Sprintf("%s, %s", errMsg, err.Error())
	}

	return nil, errors.New(errMsg)
}

func expandHomeDirAlias(path string) (string, error) {
	// Check if the path starts with the home directory alias (~).
	if !strings.HasPrefix(path, "~") {
		return path, nil // No home directory alias found, return the original path.
	}

	// On Linux, use os/user package to get the home directory.
	if os.PathSeparator == '/' {
		usr, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(usr, path[1:]), nil
	}

	// On Windows, use the os/user and os/exec packages to get the home directory.
	// Running `echo %USERPROFILE%` to get the home directory.
	cmd := exec.Command("cmd", "/C", "echo %USERPROFILE%")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	usr := strings.TrimSpace(string(output))

	return filepath.Join(usr, path[1:]), nil
}

func removeStringFromArray(target string, array []string) []string {
	var result []string

	for _, str := range array {
		if str != target {
			result = append(result, str)
		}
	}

	return result
}

func (s *Service) RequiredArchivesLoaded() bool {
	return len(s.Archives()) >= 11
}
