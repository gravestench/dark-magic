package main

import (
	"bytes"
	"encoding/binary"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

// TestPlayerArguments asserts the complete shell-free player contract, including the position of a spaced file path.
// Exact slice equality prevents future flag edits from silently changing ffplay behavior or argument boundaries.
func TestPlayerArguments(t *testing.T) {
	expected := []string{
		"-autoexit",
		"-window_title",
		"Dark Magic BIK Viewer",
		"-loglevel",
		"warning",
		"movie with spaces.bik",
	}

	arguments := playerArguments("movie with spaces.bik")
	if !reflect.DeepEqual(arguments, expected) {
		t.Fatalf("playerArguments() = %v, want %v", arguments, expected)
	}
}

// TestValidateAndDescribeBIK verifies the stable metadata format emitted before an external player is started.
// A minimal valid container keeps the test independent of proprietary game assets and native playback libraries.
func TestValidateAndDescribeBIK(t *testing.T) {
	fileName := filepath.Join(t.TempDir(), "movie.bik")
	if err := os.WriteFile(fileName, minimalValidBIK(), 0o600); err != nil {
		t.Fatalf("write BIK fixture: %v", err)
	}

	var output bytes.Buffer

	expanded, err := validateAndDescribeBIK(fileName, &output)
	if err != nil {
		t.Fatalf("validateAndDescribeBIK() error = %v", err)
	}

	if expanded != fileName {
		t.Fatalf("expanded path = %q, want %q", expanded, fileName)
	}

	const expectedOutput = "BIKi: 640x480, 1 frames, 0.033s, 0 audio track(s)\n"
	if output.String() != expectedOutput {
		t.Fatalf("metadata output = %q, want %q", output.String(), expectedOutput)
	}
}

// TestValidateAndDescribeBIKReportsReadFailure preserves the command-level context around host filesystem errors.
// The assertion avoids platform-specific PathError text while still locking the viewer's wrapping boundary.
func TestValidateAndDescribeBIKReportsReadFailure(t *testing.T) {
	fileName := filepath.Join(t.TempDir(), "missing.bik")

	_, err := validateAndDescribeBIK(fileName, &bytes.Buffer{})
	if err == nil {
		t.Fatal("validateAndDescribeBIK() error = nil, want read failure")
	}

	if !strings.HasPrefix(err.Error(), "reading BIK: ") {
		t.Fatalf("error = %q, want BIK read context", err)
	}
}

// TestPlayBIKReportsMissingPlayer verifies lookup failure occurs before command construction or file interpretation.
// An absolute nonexistent path makes the assertion independent of the developer's PATH contents.
func TestPlayBIKReportsMissingPlayer(t *testing.T) {
	playerName := filepath.Join(t.TempDir(), "missing-player-"+strconv.Itoa(os.Getpid()))

	err := playBIK(playerName, "movie.bik")
	if err == nil {
		t.Fatal("playBIK() error = nil, want executable lookup failure")
	}

	if !strings.HasPrefix(err.Error(), `locating "`+playerName+`": `) {
		t.Fatalf("error = %q, want player lookup context", err)
	}

	if !strings.HasSuffix(err.Error(), " (install FFmpeg or pass -player)") {
		t.Fatalf("error = %q, want player installation guidance", err)
	}
}

// minimalValidBIK builds the smallest Bink header accepted by the metadata decoder.
// Fixed dimensions and rate make presentation assertions deterministic without including copyrighted media.
func minimalValidBIK() []byte {
	data := make([]byte, 48)

	// Populate only decoder-required header fields; the zeroed final word is the single frame-index entry.
	copy(data[:4], "BIKi")
	binary.LittleEndian.PutUint32(data[4:8], uint32(len(data)-8))
	binary.LittleEndian.PutUint32(data[8:12], 1)
	binary.LittleEndian.PutUint32(data[16:20], 1)
	binary.LittleEndian.PutUint32(data[20:24], 640)
	binary.LittleEndian.PutUint32(data[24:28], 480)
	binary.LittleEndian.PutUint32(data[28:32], 30)
	binary.LittleEndian.PutUint32(data[32:36], 1)

	return data
}
