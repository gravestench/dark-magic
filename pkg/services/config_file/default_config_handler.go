package config_file

import (
	"fmt"

	"github.com/gravestench/runtime"
)

func (s *Service) applyDefaultConfig(candidate runtime.S) error {
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
	cfgPath := prefixIfPathRelative(s.ConfigDirectory(), target.ConfigFilePath())
	cfgDefault := target.DefaultConfig()
	cfgCurrent, err := s.GetConfig(cfgPath)

	if err != nil || cfgCurrent == nil {
		cfgCurrent, err = s.CreateConfig(cfgPath)
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

	s.log.Info().Msgf("config file for %q service can be found at: %v", name, s.GetPath(target.ConfigFilePath()))

	return s.SaveConfig(cfgPath)
}
