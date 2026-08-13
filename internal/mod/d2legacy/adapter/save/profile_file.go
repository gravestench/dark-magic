package save

import (
	"fmt"
	"os"
	"path/filepath"
)

func WriteProfileFile(path string, profile Profile) error {
	if path == "" {
		return fmt.Errorf("%w: path is required", ErrProfile)
	}
	directory := filepath.Dir(path)
	if err := os.MkdirAll(directory, 0o700); err != nil {
		return fmt.Errorf("%w: create profile directory: %v", ErrProfile, err)
	}
	temporary, err := os.CreateTemp(directory, ".dark-magic-profile-*")
	if err != nil {
		return fmt.Errorf("%w: create temporary file: %v", ErrProfile, err)
	}
	temporaryPath := temporary.Name()
	defer os.Remove(temporaryPath)
	if err := temporary.Chmod(0o600); err != nil {
		_ = temporary.Close()
		return fmt.Errorf("%w: protect temporary file: %v", ErrProfile, err)
	}
	if err := EncodeProfile(temporary, profile); err != nil {
		_ = temporary.Close()
		return err
	}
	if err := temporary.Sync(); err != nil {
		_ = temporary.Close()
		return fmt.Errorf("%w: sync temporary file: %v", ErrProfile, err)
	}
	if err := temporary.Close(); err != nil {
		return fmt.Errorf("%w: close temporary file: %v", ErrProfile, err)
	}
	if err := os.Rename(temporaryPath, path); err != nil {
		return fmt.Errorf("%w: replace profile: %v", ErrProfile, err)
	}
	directoryHandle, err := os.Open(directory)
	if err != nil {
		return fmt.Errorf("%w: open profile directory: %v", ErrProfile, err)
	}
	defer directoryHandle.Close()
	if err := directoryHandle.Sync(); err != nil {
		return fmt.Errorf("%w: sync profile directory: %v", ErrProfile, err)
	}
	return nil
}

func ReadProfileFile(path string) (Profile, error) {
	if path == "" {
		return Profile{}, fmt.Errorf("%w: path is required", ErrProfile)
	}
	file, err := os.Open(path)
	if err != nil {
		return Profile{}, fmt.Errorf("%w: open profile: %w", ErrProfile, err)
	}
	defer file.Close()
	return DecodeProfile(file)
}
