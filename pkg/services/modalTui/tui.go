package modalTui

import (
	"fmt"
	"sort"
)

func (s *Service) CurrentModal() string {
	return s.selectedModal
}

func (s *Service) GetModalNames() (keys []string) {
	return s.getSortedModalNameList()
}

func (s *Service) SwitchModal(name string) error {
	if _, exists := s.modals[name]; !exists {
		return fmt.Errorf("modal %q does not exist", name)
	}

	s.selectedModal = name

	return nil
}

func (s *Service) getSortedModalNameList() (keys []string) {
	for key := range s.modals {
		keys = append(keys, key)
	}

	sort.Strings(keys)

	return
}
