package webRouter

import (
	"time"

	"github.com/gin-gonic/gin"
	"k8s.io/utils/strings/slices"

	"github.com/gravestench/servicemesh"

	"github.com/gravestench/dark-magic/pkg/services/webRouter/middleware/logger"
)

func (s *Service) RouteRoot() *gin.Engine {
	return s.root
}

func (s *Service) Reload() {
	if time.Since(s.reloadDebounce) < time.Second {
		return
	}

	s.reloadDebounce = time.Now()

	s.log.Warn("reloading")

	if s.boundServices == nil {
		s.log.Info("initializing routes")
	} else {
		s.log.Info("re-initializing routes")
	}

	// forget any already-bound services
	s.boundServices = nil
	s.boundServices = make(map[string]*struct{})

	mode := gin.ReleaseMode
	if s.config.debug {
		mode = gin.DebugMode
	}

	s.root = nil

	gin.SetMode(mode)

	// set up the root and protected route groups
	s.root = gin.New()
	s.root.RemoveExtraSlash = true

	// IMPORTANT, add the common middleware needs to be added before
	// creating the protected route group, otherwise it wont have the
	// middleware
	s.root.Use(func(c *gin.Context) {
		logger.Logger("gin", s.Logger())(c)
	})
}

func (s *Service) beginDynamicRouteBinding(mesh servicemesh.Mesh) {
	for {
		if s.shouldInit(mesh) {
			s.Reload()
		}

		s.bindNewRoutes(mesh)
		time.Sleep(time.Second)
	}
}

func (s *Service) shouldInit(mesh servicemesh.Mesh) bool {
	if s.boundServices == nil {
		return true // base case, happens at app start
	}

	if s.root == nil {
		return true // base case, happens at app start
	}

	// in the event that a service is removed by the
	// service manager for whatever reason, we need to check
	// if that was something that had routes. if it was, we need
	// to re-init the router (we can't actually remove routes in gin)

	// we will check if any of the services that have routes are no longer
	// in the list of the service managers services
	allExistingServiceNames := make([]string, 0)
	for _, candidate := range mesh.Services() {
		if svc, ok := candidate.(IsRouteInitializer); ok {
			allExistingServiceNames = append(allExistingServiceNames, svc.Name())
			continue
		}
	}

	// iterate over each bound service, check if its name
	// exists as a substring inside of our lookup string
	for key, _ := range s.boundServices {
		if !slices.Contains(allExistingServiceNames, key) {
			return true
		}
	}

	return false
}

func (s *Service) bindNewRoutes(mesh servicemesh.Mesh) {
	for _, candidate := range mesh.Services() {
		svcToInit, ok := candidate.(servicemesh.Service)
		if !ok {
			continue
		}

		if svc, ok := candidate.(servicemesh.HasDependencies); ok {
			if !svc.DependenciesResolved() {
				continue
			}
		}

		if _, alreadyBound := s.boundServices[svcToInit.Name()]; alreadyBound {
			continue
		}

		groupPrefix := ""
		if svc, ok := candidate.(HasRouteSlug); ok {
			groupPrefix = svc.Slug()
		}

		defer func() {
			_ = recover() // dont worry about this lol
		}()

		// handle route init
		if r, ok := candidate.(IsRouteInitializer); ok {
			r.InitRoutes(s.root.Group(groupPrefix))
			s.boundServices[svcToInit.Name()] = nil // make 0-size entry
			s.log.Info("binding routes", "for", svcToInit.Name())

			continue
		}
	}
}
