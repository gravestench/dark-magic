package webRouter

import (
	"log/slog"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gravestench/servicemesh"

	"github.com/gravestench/dark-magic/pkg/services/webRouter/middleware/static_assets"
)

type Service struct {
	log    *slog.Logger
	config *Config

	root *gin.Engine

	boundServices map[string]*struct{} // holds 0-size entries

	reloadDebounce time.Time
	mtx            sync.Mutex
}

func (s *Service) SetLogger(l *slog.Logger) {
	s.log = l
}

func (s *Service) Logger() *slog.Logger {
	return s.log
}

func (s *Service) Init(mesh servicemesh.Mesh) {
	gin.SetMode("release")
	mesh.Add(&static_assets.Middleware{})
	s.root = gin.New()
	//go s.beginDynamicRouteBinding(mesh)

	for _, service := range mesh.Services() {
		s.initRoutesForService(service)
	}
}

func (s *Service) Name() string {
	return "Web Router"
}

func (s *Service) Ready() bool {
	if s.config == nil {
		return false
	}

	return true
}

func (s *Service) initRoutesForService(service servicemesh.Service) {
	candidate, ok := service.(IsRouteInitializer)
	if !ok {
		return
	}

	groupPrefix := ""
	if svc, ok := candidate.(HasRouteSlug); ok {
		groupPrefix = svc.Slug()
	}

	for s.RouteRoot() == nil {
		time.Sleep(time.Second)
	}

	s.mtx.Lock()
	defer s.mtx.Unlock()

	if _, bound := s.boundServices[candidate.Name()]; bound {
		return
	}

	defer func() {
		if err := recover(); err != nil {
			s.Logger().Warn("binding routes", "error", err)
		}
	}()

	candidate.InitRoutes(s.RouteRoot().Group(groupPrefix))
}
