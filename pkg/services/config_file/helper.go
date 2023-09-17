package config_file

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"time"
)

func directoryExists(path string) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		// Check if the error indicates that the directory does not exist
		if os.IsNotExist(err) {
			return false, nil
		}
		// Return the error for any other issue
		return false, err
	}

	// Check if the given path is a directory
	if !info.IsDir() {
		return false, fmt.Errorf("%s is not a directory", path)
	}

	return true, nil
}

func directoryExistsForFile(p string) (bool, error) {
	dir := filepath.Dir(p)

	info, err := os.Stat(dir)
	if err != nil {
		// Check if the error indicates that the directory does not exist
		if os.IsNotExist(err) {
			return false, nil
		}
		// Return the error for any other issue
		return false, err
	}

	// Check if the given path is a directory
	if !info.IsDir() {
		return false, fmt.Errorf("%s is not a directory", dir)
	}

	return true, nil
}

func pathExists(p string) (bool, error) {
	_, err := os.Stat(p)
	if err != nil {
		// Check if the error indicates that the directory does not exist
		if os.IsNotExist(err) {
			return false, nil
		}
		// Return the error for any other issue
		return false, err
	}

	return true, nil
}

func prefixIfPathRelative(dir, path string) string {
	dir = expandHomeDirectory(dir)

	if !filepath.IsAbs(path) {
		return filepath.Join(dir, path)
	}

	return path
}

func expandHomeDirectory(path string) string {
	if strings.HasPrefix(path, "~") {
		homeDir, err := os.UserHomeDir()
		if err == nil {
			path = strings.Replace(path, "~", homeDir, 1)
		}
	}
	return path
}

func compressPath(filepath string) string {
	usr, err := user.Current()
	if err != nil {
		return filepath
	}
	homeDir := usr.HomeDir

	if strings.HasPrefix(filepath, homeDir) {
		filepath = strings.Replace(filepath, homeDir, "~", 1)
	}

	return filepath
}

func touch(path string) error {
	info, err := os.Stat(path)

	if err != nil {
		// Path does not exist, create it
		if os.IsNotExist(err) {
			err := os.MkdirAll(path, 0755)
			if err != nil {
				return err
			}
		} else {
			return err
		}
	} else {
		// Path exists, update its access and modification times
		currentTime := time.Now()

		// Update directory
		if info.IsDir() {
			err := os.Chtimes(path, currentTime, currentTime)
			if err != nil {
				return err
			}
		} else { // Update file
			err := os.Chtimes(path, currentTime, currentTime)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
