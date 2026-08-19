package capture

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"image"
	_ "image/png"
	"os"
	"strings"
)

const (
	standardVisibilityDivisor = 50
	sparseVisibilityDivisor   = 1000
)

var errBlankFrame = errors.New("framebuffer contains no visible pixels")

// Result describes one verified screenshot without exposing the local output directory in the report format.
type Result struct {
	Scene  string `json:"scene"`
	File   string `json:"file"`
	SHA256 string `json:"sha256"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

// Report is the versioned, deterministic manifest consumed by local visual-review tooling.
type Report struct {
	Version int      `json:"version"`
	Results []Result `json:"results"`
}

// screenshotFilename preserves capture order in lexical listings while keeping scene names filesystem-safe.
func screenshotFilename(sequence int, scene string) string {
	return fmt.Sprintf("%02d-%s.png", sequence, safeName(scene))
}

// inspectScreenshot verifies decoded visibility before publishing dimensions and a digest of the exact file bytes.
func inspectScreenshot(path, scene, name string) (Result, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Result{}, err
	}

	file, err := os.Open(path)
	if err != nil {
		return Result{}, err
	}

	frame, _, decodeErr := image.Decode(file)
	closeErr := file.Close()
	// A decode failure describes the artifact itself and therefore keeps precedence over a secondary close failure.
	if decodeErr != nil {
		return Result{}, decodeErr
	}

	if closeErr != nil {
		return Result{}, closeErr
	}

	if !hasVisiblePixels(frame, scene) {
		return Result{}, errBlankFrame
	}

	digest := sha256.Sum256(data)
	bounds := frame.Bounds()

	return Result{
		Scene:  scene,
		File:   name,
		SHA256: hex.EncodeToString(digest[:]),
		Width:  bounds.Dx(),
		Height: bounds.Dy(),
	}, nil
}

// hasVisiblePixels rejects blank framebuffers while allowing known text-heavy overlays a deliberately sparse budget.
func hasVisiblePixels(frame image.Image, scene string) bool {
	bounds := frame.Bounds()
	requiredPixels := requiredVisiblePixels(bounds, scene)
	visiblePixels := 0

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			red, green, blue, alpha := frame.At(x, y).RGBA()
			if alpha == 0 || red <= 0x0400 && green <= 0x0400 && blue <= 0x0400 {
				continue
			}

			visiblePixels++
			if visiblePixels >= requiredPixels {
				return true
			}
		}
	}

	return false
}

// requiredVisiblePixels gives sparse authentic overlays a lower threshold without weakening ordinary scene checks.
func requiredVisiblePixels(bounds image.Rectangle, scene string) int {
	pixelCount := bounds.Dx() * bounds.Dy()
	if sceneAllowsSparseFrame(scene) {
		// These overlays contain only text or a small transition over the world. Direct boot has no world beneath them,
		// so the ordinary two-percent threshold would incorrectly reject authentic but sparse frames.
		return (pixelCount + sparseVisibilityDivisor - 1) / sparseVisibilityDivisor
	}

	return (pixelCount + standardVisibilityDivisor - 1) / standardVisibilityDivisor
}

// sceneAllowsSparseFrame names the compatibility set whose direct-boot presentation lacks a full world backdrop.
func sceneAllowsSparseFrame(scene string) bool {
	switch scene {
	case "death", "game_loading", "npc_dialogue", "ground_items", "chat", "overhead_labels":
		return true
	default:
		return false
	}
}

// safeName normalizes artifact components without changing the established lowercase and replacement rules.
func safeName(value string) string {
	value = strings.Map(safeNameRune, strings.ToLower(value))

	return strings.Trim(value, "-")
}

// safeNameRune retains portable filename characters and replaces every other rune with one hyphen.
func safeNameRune(current rune) rune {
	if current >= 'a' && current <= 'z' ||
		current >= '0' && current <= '9' ||
		current == '-' ||
		current == '_' {
		return current
	}

	return '-'
}
