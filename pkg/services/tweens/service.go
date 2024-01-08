package tweens

import (
	"time"

	"github.com/gravestench/servicemesh"

	"github.com/gravestench/dark-magic/pkg/services/common"
)

type Service struct {
	common.Service
	queue      []*Tween
	lastUpdate time.Time
	*Config
}

func (s *Service) DependenciesResolved() bool {
	if s.Config == nil {
		return false
	}

	return true
}

func (s *Service) ResolveDependencies(services []servicemesh.Service) {
	// our only dependency is the config, which is handled by another service
	return
}

func (s *Service) Init(mesh servicemesh.Mesh) {
	s.lastUpdate = time.Now()
	go s.beginUpdating()
}

func (s *Service) beginUpdating() {
	ticker := time.NewTicker(time.Second / time.Duration(s.Config.TickRate))

	for <-ticker.C; ; {
		for idx := range s.queue {
			delta := time.Since(s.lastUpdate)
			s.queue[idx].Update(delta)
		}
	}
}

func (s *Service) Name() string {
	return "Tweens"
}

func (s *Service) Ready() bool {
	return s.Config != nil
}

// New creates a new tween, but does not add it for processing.
func (s *Service) New() *Tween {
	t := &Tween{}
	t.tweenConfig = &tweenConfig{}

	t.justStarted = true
	t.Time(defaultDuration)
	t.Ease(defaultEase)
	t.OnStart(func() {})
	t.OnComplete(func() {})
	t.OnUpdate(func(float64) {})

	return t
}

// Add the given teen to the processing queue
func (s *Service) Add(t *Tween) {
	s.queue = append(s.queue, t)
}

// Remove the given tween from the queue
func (s *Service) Remove(t *Tween) {
	for idx := range s.queue {
		if s.queue[idx] != t {
			continue
		}

		s.queue = append(s.queue[:idx], s.queue[idx+1:]...)

		break
	}
}
