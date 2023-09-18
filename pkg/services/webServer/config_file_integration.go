package webServer

import (
	"github.com/gravestench/dark-magic/pkg/services/configFile"
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

func (s *Service) ConfigFileName() string {
	return "web_server.json"
}

func (s *Service) Config() (*configFile.Config, error) {
	return s.cfgManager.GetConfigByFileName(s.ConfigFileName())
}

func (s *Service) DefaultConfig() (cfg configFile.Config) {
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
