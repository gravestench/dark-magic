package lua

import (
	"os"
	"time"
)

func (s *Service) runScript(script string) {
	if err := s.state.DoFile(script); err != nil {
		s.logger.Warn().Msgf("executing init script %q: %v", script, err)
	}
}

func (s *Service) watchFileAndCallOnChange(filePath string, onChange func()) {
	// Get the initial modification time of the file.
	initialModTime, err := getFileModTime(filePath)
	if err != nil {
		s.logger.Warn().Msgf("getting init script mod time:", err)
		return
	}

	for {
		// Sleep for a short duration before checking the file again.
		time.Sleep(1 * time.Second)

		// Get the current modification time of the file.
		currentModTime, err := getFileModTime(filePath)
		if err != nil {
			s.logger.Warn().Msgf("getting init script mod time:", err)
			continue
		}

		// Compare the current modification time with the initial one.
		if currentModTime.After(initialModTime) {
			onChange()
			// Update the initial modification time to the current one.
			initialModTime = currentModTime
		}
	}
}

func getFileModTime(filePath string) (time.Time, error) {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return time.Time{}, err
	}
	return fileInfo.ModTime(), nil
}
