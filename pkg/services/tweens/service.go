package tweens

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gravestench/servicemesh"
	"github.com/yuin/gopher-lua"

	"github.com/gravestench/dark-magic/pkg/services/common"
	"github.com/gravestench/dark-magic/pkg/services/luaManager"
)

type Service struct {
	common.Service
	queue      []*Tween
	lastUpdate time.Time
	mux        sync.Mutex
	cancel     context.CancelFunc
	lua        luaManager.Dependency
	*Config
}

var _ servicemesh.HasGracefulShutdown = &Service{}

func (s *Service) DependenciesResolved() bool {
	if s.Config == nil {
		return false
	}
	if s.lua == nil {
		return false
	}

	return true
}

func (s *Service) ResolveDependencies(services []servicemesh.Service) {
	for _, service := range services {
		if candidate, ok := service.(luaManager.Dependency); ok {
			s.lua = candidate
		}
	}
}

func (s *Service) Init(mesh servicemesh.Mesh) {
	s.lastUpdate = time.Now()
	ctx, cancel := context.WithCancel(context.Background())
	s.cancel = cancel
	go s.beginUpdating(ctx)
}

func (s *Service) beginUpdating(ctx context.Context) {
	ticker := time.NewTicker(time.Second / time.Duration(s.Config.TickRate))
	defer ticker.Stop()

	for {
		select {
		case now := <-ticker.C:
			delta := now.Sub(s.lastUpdate)
			s.lastUpdate = now
			s.mux.Lock()
			queue := append([]*Tween(nil), s.queue...)
			s.mux.Unlock()
			for _, tween := range queue {
				if s.lua == nil {
					tween.Update(delta)
					continue
				}
				_ = s.lua.WithState(func(_ *lua.LState) error {
					tween.Update(delta)
					return nil
				})
			}
		case <-ctx.Done():
			return
		}
	}
}

func (s *Service) OnShutdown() {
	if s.cancel != nil {
		s.cancel()
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
	tween := &Tween{}

	tween.id = uuid.New()
	tween.justStarted = true
	tween.Time(defaultDuration)
	tween.Ease(defaultEase)
	tween.OnStart(func() {})
	tween.OnComplete(func() {})
	tween.OnUpdate(func(float64) {})

	return tween
}

// Add the given tween to the processing queue
func (s *Service) Add(t *Tween) {
	s.mux.Lock()
	defer s.mux.Unlock()
	s.queue = append(s.queue, t)
}

// Remove the given tween from the queue
func (s *Service) Remove(t *Tween) {
	s.mux.Lock()
	defer s.mux.Unlock()
	for idx := range s.queue {
		if s.queue[idx] != t {
			continue
		}

		s.queue = append(s.queue[:idx], s.queue[idx+1:]...)

		break
	}
}
