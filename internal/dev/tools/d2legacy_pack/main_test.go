package main

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/gravestench/dark-magic/internal/content"
)

// TestWriteArchiveFileCreatesExpectedArchive protects parent creation and the archive's deterministic byte format.
func TestWriteArchiveFileCreatesExpectedArchive(t *testing.T) {
	outputPath := filepath.Join(t.TempDir(), "nested", "d2legacy.zip")
	if err := writeArchiveFile(outputPath); err != nil {
		t.Fatal(err)
	}

	actual, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatal(err)
	}

	var expected bytes.Buffer
	if err := content.WriteD2LegacyArchive(&expected); err != nil {
		t.Fatal(err)
	}

	// Exact bytes protect entry ordering and ZIP metadata as well as the embedded file contents.
	if !bytes.Equal(actual, expected.Bytes()) {
		t.Fatalf("archive bytes differ: got %d bytes, want %d", len(actual), expected.Len())
	}
}

// TestWriteArchiveFilePreservesPathErrors ensures invalid aliases fail before the command creates output state.
func TestWriteArchiveFilePreservesPathErrors(t *testing.T) {
	const outputName = "~someone/d2legacy.zip"

	err := writeArchiveFile(outputName)
	if err == nil {
		t.Fatal("writeArchiveFile succeeded with a named home directory")
	}

	if want := `paths: named home directories are unsupported in "~someone/d2legacy.zip"`; err.Error() != want {
		t.Fatalf("writeArchiveFile error = %q, want %q", err, want)
	}
}

// TestWriteAndCloseArchivePrefersWriteFailure verifies cleanup cannot hide the failure that made an archive invalid.
func TestWriteAndCloseArchivePrefersWriteFailure(t *testing.T) {
	writeErr := errors.New("write failed")
	closeErr := errors.New("close failed")
	destination := &controlledArchiveDestination{writeErr: writeErr, closeErr: closeErr}

	err := writeAndCloseArchive(destination)
	if !errors.Is(err, writeErr) {
		t.Fatalf("writeAndCloseArchive error = %v, want write failure", err)
	}

	if errors.Is(err, closeErr) {
		t.Fatalf("writeAndCloseArchive error = %v, cleanup failure replaced write failure", err)
	}

	if !destination.closed {
		t.Fatal("writeAndCloseArchive did not close the destination after a write failure")
	}
}

// TestWriteAndCloseArchiveReturnsCloseFailure ensures a complete archive still reports a failed final file close.
func TestWriteAndCloseArchiveReturnsCloseFailure(t *testing.T) {
	closeErr := errors.New("close failed")
	destination := &controlledArchiveDestination{closeErr: closeErr}

	err := writeAndCloseArchive(destination)
	if !errors.Is(err, closeErr) {
		t.Fatalf("writeAndCloseArchive error = %v, want close failure", err)
	}

	if !destination.closed {
		t.Fatal("writeAndCloseArchive did not close the destination after a successful write")
	}

	if destination.Len() == 0 {
		t.Fatal("writeAndCloseArchive produced an empty archive before the close failure")
	}
}

// controlledArchiveDestination records cleanup and can fail either operation to expose error-precedence guarantees.
type controlledArchiveDestination struct {
	bytes.Buffer
	writeErr error
	closeErr error
	closed   bool
}

// Write either records archive bytes or injects the configured write failure without retaining partial output.
func (destination *controlledArchiveDestination) Write(data []byte) (int, error) {
	if destination.writeErr != nil {
		return 0, destination.writeErr
	}

	return destination.Buffer.Write(data)
}

// Close records the ownership handoff before returning the configured cleanup result.
func (destination *controlledArchiveDestination) Close() error {
	destination.closed = true
	return destination.closeErr
}
