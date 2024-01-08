package assetLoader

import (
	"encoding/json"
	"fmt"

	"github.com/gravestench/dark-magic/pkg/services/configManager"
)

type Config struct {
	Dc6CacheMB  int
	DccCacheMB  int
	Ds1CacheMB  int
	Dt1CacheMB  int
	CofCacheMB  int
	FontCacheMB int
	Pl2CacheMB  int
	TblCacheMB  int
	TsvCacheMB  int
	WavCacheMB  int
}

var _ configManager.HasConfiguration = &Service{}

func (s *Service) ConfigFileName() string {
	return "asset_loader.json"
}

func (s *Service) DefaultConfigData() (data []byte) {
	data, _ = json.MarshalIndent(&Config{
		Dc6CacheMB:  20,
		DccCacheMB:  20,
		Ds1CacheMB:  20,
		Dt1CacheMB:  20,
		CofCacheMB:  20,
		FontCacheMB: 20,
		Pl2CacheMB:  20,
		TblCacheMB:  20,
		TsvCacheMB:  20,
		WavCacheMB:  50,
	}, "", "\t")

	return data
}

func (s *Service) IngestConfig(handle *configManager.ConfigHandle) error {
	data, err := handle.Data()
	if err != nil {
		return fmt.Errorf("getting data from config handle: %v", err)
	}

	if err = json.Unmarshal(data, &s.Config); err != nil {
		return err
	}

	return nil
}
