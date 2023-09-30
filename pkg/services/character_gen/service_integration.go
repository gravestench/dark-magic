package character_gen

import (
	"fmt"
	"os"
)

// FileLogger represents a logger that logs character information to a file.
type FileLogger struct {
	logFile *os.File
}

// NewFileLogger creates a new FileLogger instance and opens a log file for writing.
func NewFileLogger(logFileName string) (*FileLogger, error) {
	logFile, err := os.Create(logFileName)
	if err != nil {
		return nil, err
	}
	return &FileLogger{
		logFile: logFile,
	}, nil
}

// LogCharacterAttributes logs the attributes of a character to the file.
func (fl *FileLogger) LogCharacterAttributes(character *Character) error {
	_, err := fmt.Fprintf(fl.logFile, "Name: %s, Strength: %d, Dexterity: %d, Vitality: %d, Energy: %d\n", character.Name, character.Attributes.Strength, character.Attributes.Dexterity, character.Attributes.Vitality, character.Attributes.Energy)
	if err != nil {
		return err
	}
	return nil
}

// Close closes the log file.
func (fl *FileLogger) Close() error {
	if fl.logFile != nil {
		return fl.logFile.Close()
	}
	return nil
}
