package screenManager

// Modes returns a list of strings which are the mode ID's for each modal
func (s *Service) Modes() (modes []string) {
	for _, modal := range s.modals {
		modes = append(modes, modal.Mode())
	}

	return
}

// Mode returns the current mode string
func (s *Service) Mode() string {
	return s.currentMode
}

// SetMode sets the current mode to be rendered
func (s *Service) SetMode(mode string) {
	modal := s.getMode(mode)
	if modal == nil {
		return
	}

	for _, child := range s.rootNode.Children() {
		child.Disable()
	}

	modal.Renderable().Enable()
}

func (s *Service) getMode(mode string) ModalGameUI {
	for _, modal := range s.modals {
		if mode == modal.Mode() {
			return modal
		}
	}

	return nil
}
