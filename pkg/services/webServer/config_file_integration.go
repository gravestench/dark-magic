package webServer

import (
	"encoding/json"
	"fmt"

	"github.com/gravestench/dark-magic/pkg/services/configManager"
)

type Config struct {
	Port         int
	Tls          bool
	AutoCert     bool
	CertFilepath string
	KeyFilepath  string
}

var _ configManager.HasConfiguration = &Service{}

func (s *Service) ConfigFileName() string {
	return "web_server.json"
}

func (s *Service) DefaultConfigData() []byte {
	var cfg Config

	cfg.Port = 8080
	cfg.Tls = false
	cfg.AutoCert = false
	cfg.CertFilepath = "cert.pem"
	cfg.KeyFilepath = "cert.key"

	data, _ := json.MarshalIndent(&cfg, "", "\t")

	return data
}

func (s *Service) IngestConfig(handle *configManager.ConfigHandle) error {
	data, err := handle.Data()
	if err != nil {
		return fmt.Errorf("getting config data: %v", err)
	}

	if err = json.Unmarshal(data, &s.config); err != nil {
		return fmt.Errorf("unmarshalling config data: %v", err)
	}

	return nil
}
