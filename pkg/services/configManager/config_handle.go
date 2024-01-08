package configManager

import (
	"os"
)

// ConfigHandle represents a config file which is being
// managed by the config manager service. Byte data has
// a getter/setter, and there is a getter for the filepath.
type ConfigHandle struct {
	manager  *Service
	filepath string
}

func (h *ConfigHandle) FilePath() string {
	return h.filepath
}

func (h *ConfigHandle) Data() ([]byte, error) {
	return os.ReadFile(h.filepath)
}

func (h *ConfigHandle) SetData(data []byte) error {
	return os.WriteFile(h.filepath, data, 0o755)
}
