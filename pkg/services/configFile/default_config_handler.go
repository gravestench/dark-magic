package configFile

import (
	"fmt"

	"github.com/gravestench/servicemesh"
)

func (s *Service) initConfigForServiceCandidate(candidate servicemesh.Service) error {
	s.mux.Lock()
	defer s.mux.Unlock()

	// check if the service does not have defaults
	target, ok := candidate.(HasDefaultConfig)
	if !ok {
		return nil
	}

	// check if we already know about it
	name := candidate.Name()
	if _, found := s.servicesWithDefaultConfigs[name]; found {
		return nil
	}

	// add it to our list
	s.servicesWithDefaultConfigs[name] = target

	// get the current and default configs
	cfgPath := prefixIfPathRelative(s.ConfigDirectory(), target.ConfigFileName())
	cfgDefault := target.DefaultConfig()
	cfgCurrent, err := s.getConfigUnsafe(cfgPath)

	if err != nil || cfgCurrent == nil {
		cfgCurrent, err = s.createConfigUnsafe(cfgPath)
		if err != nil {
			return fmt.Errorf("creating config %q: %v", cfgPath, err)
		}
	}

	for groupKey, group := range cfgDefault.groups {
		for key, defaultValue := range group {
			currentGroup := cfgCurrent.Group(groupKey)
			currentGroup.SetDefault(key, defaultValue)
		}
	}

	s.log.Info("default config applied", "for", name, "path", s.GetFilePath(target.ConfigFileName()))

	return s.saveConfigUnsafe(cfgPath)
}
