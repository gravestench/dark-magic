package backgroundMusic

import (
	"github.com/gravestench/servicemesh"
)

type BackgroundMusicManager interface {
	servicemesh.Service
	servicemesh.HasDependencies
	servicemesh.HasLogger
	SetBackgroundMusic(path string) error
}
