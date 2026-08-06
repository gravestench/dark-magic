// Package capture records local visual-review artifacts without embedding game
// imagery in tests or source control.
package capture

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"image"
	_ "image/png"
	"os"
	"path/filepath"
	"strings"
)

type Screenshotter interface {
	CaptureScreenshot(string) error
}

type Result struct {
	Scene  string `json:"scene"`
	File   string `json:"file"`
	SHA256 string `json:"sha256"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

type Report struct {
	Version int      `json:"version"`
	Results []Result `json:"results"`
}

// Session captures the first stable frame for each requested scene.
type Session struct {
	directory string
	wanted    map[string]bool
	captured  map[string]bool
	settle    int
	current   string
	frames    int
	results   []Result
	capturer  Screenshotter
	err       error
}

func New(directory, scenes string, settleFrames int, capturer Screenshotter) (*Session, error) {
	if directory == "" || capturer == nil {
		return nil, fmt.Errorf("capture: directory and screenshotter are required")
	}
	if settleFrames < 1 {
		return nil, fmt.Errorf("capture: settle frames must be positive")
	}
	wanted := make(map[string]bool)
	for _, scene := range strings.Split(scenes, ",") {
		scene = strings.TrimSpace(scene)
		if scene != "" {
			wanted[scene] = true
		}
	}
	if len(wanted) == 0 {
		return nil, fmt.Errorf("capture: at least one scene is required")
	}
	if err := os.MkdirAll(directory, 0o755); err != nil {
		return nil, err
	}
	return &Session{directory: directory, wanted: wanted, captured: make(map[string]bool), settle: settleFrames, capturer: capturer}, nil
}

func (s *Session) Observe(stack []string) {
	if s.err != nil || len(stack) == 0 {
		return
	}
	scene := stack[len(stack)-1]
	if scene != s.current {
		s.current, s.frames = scene, 0
	}
	s.frames++
	if !s.wanted[scene] || s.captured[scene] || s.frames < s.settle {
		return
	}
	name := fmt.Sprintf("%02d-%s.png", len(s.results)+1, safeName(scene))
	path := filepath.Join(s.directory, name)
	if err := s.capturer.CaptureScreenshot(path); err != nil {
		s.err = err
		return
	}
	result, err := inspect(path, scene, name)
	if err != nil {
		s.err = err
		return
	}
	s.captured[scene] = true
	s.results = append(s.results, result)
}

func (s *Session) Close() error {
	report := Report{Version: 1, Results: append([]Result(nil), s.results...)}
	data, err := json.MarshalIndent(report, "", "  ")
	if err == nil {
		data = append(data, '\n')
		err = os.WriteFile(filepath.Join(s.directory, "report.json"), data, 0o644)
	}
	if s.err != nil {
		return fmt.Errorf("capture: %w", s.err)
	}
	return err
}

func inspect(path, scene, name string) (Result, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Result{}, err
	}
	file, err := os.Open(path)
	if err != nil {
		return Result{}, err
	}
	configuration, _, decodeErr := image.DecodeConfig(file)
	closeErr := file.Close()
	if decodeErr != nil {
		return Result{}, decodeErr
	}
	if closeErr != nil {
		return Result{}, closeErr
	}
	digest := sha256.Sum256(data)
	return Result{Scene: scene, File: name, SHA256: hex.EncodeToString(digest[:]), Width: configuration.Width, Height: configuration.Height}, nil
}

func safeName(value string) string {
	value = strings.Map(func(current rune) rune {
		if current >= 'a' && current <= 'z' || current >= '0' && current <= '9' || current == '-' || current == '_' {
			return current
		}
		return '-'
	}, strings.ToLower(value))
	return strings.Trim(value, "-")
}
