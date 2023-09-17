package static_assets

import (
	"time"

	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"

	"github.com/gravestench/runtime"
)

type IsWebRouter interface {
	RouteRoot() *gin.Engine
	Reload()
}

type Middleware struct {
	log        *zerolog.Logger
	router     IsWebRouter
	lastRouter interface{}
}

func (m *Middleware) Name() string {
	return "Static Assets Middleware"
}

func (m *Middleware) Init(rt runtime.R) {
	m.initMiddleware()

	for {
		if m.isRouterChanged() {
			m.log.Warn().Msg("re-initializing")
			m.initMiddleware()
		}

		time.Sleep(time.Millisecond * 100)
	}
}

func (m *Middleware) initMiddleware() {
	r := m.router.RouteRoot()
	m.log.Info().Msg("setting up routes for static assets")
	r.Use(gzip.Gzip(gzip.DefaultCompression))
	r.NoRoute(m.staticWebUIHandler)
}

func (m *Middleware) Logger() *zerolog.Logger {
	return m.log
}

func (m *Middleware) BindLogger(logger *zerolog.Logger) {
	m.log = logger
}
