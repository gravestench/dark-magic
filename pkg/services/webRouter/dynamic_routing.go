package webRouter

import (
	"time"

	"github.com/gin-gonic/gin"
	"k8s.io/utils/strings/slices"

	"github.com/gravestench/runtime"

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

	s.log.Warn().Msg("reloading")

	if s.boundServices == nil {
		s.log.Info().Msg("initializing routes")
	} else {
		s.log.Info().Msg("re-initializing routes")
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
	s.root.Use(logger.Logger("gin", s.log))
}

func (s *Service) beginDynamicRouteBinding(rt runtime.R) {
	for {
		if s.shouldInit(rt) {
			s.Reload()
		}

		s.bindNewRoutes(rt)
		time.Sleep(time.Second)
	}
}

func (s *Service) shouldInit(rt runtime.R) bool {
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
	for _, candidate := range rt.Services() {
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

func (s *Service) bindNewRoutes(rt runtime.R) {
	for _, candidate := range rt.Services() {
		svcToInit, ok := candidate.(runtime.Service)
		if !ok {
			continue
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
			s.log.Info().Msgf("binding routes for the %q service", svcToInit.Name())

			continue
		}
	}
}
