package web_server

import (
	"github.com/gravestench/runtime/examples/services/config_file"
)

const (
	keyPort     = "port"
	defaultPort = 8080

	keyTls     = "tls"
	defaultTls = false

	keyAutocert     = "autocert"
	defaultAutocert = false

	keyCertFilepath     = "tls.CertFilepath"
	defaultCertFilepath = "cert.pem"

	keyKeyFilepath     = "tls.KeyFilepath"
	defaultKeyFilepath = "cert.key"
)

func (s *Service) ConfigFilePath() string {
	return "web_server.json"
}

func (s *Service) Config() (*config_file.Config, error) {
	return s.cfgManager.GetConfig(s.ConfigFilePath())
}

func (s *Service) DefaultConfig() (cfg config_file.Config) {
	g := cfg.Group("Web Server")

	for key, val := range map[string]any{
		keyPort:         defaultPort,
		keyTls:          defaultTls,
		keyAutocert:     defaultAutocert,
		keyCertFilepath: defaultCertFilepath,
		keyKeyFilepath:  defaultKeyFilepath,
	} {
		g.Set(key, val)
	}

	return
}
