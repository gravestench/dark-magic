package locale

import (
	"github.com/gravestench/servicemesh"

	"github.com/gravestench/dark-magic/pkg/services/webRouter"
)

var (
	_ servicemesh.Service          = &Service{}
	_ servicemesh.HasLogger        = &Service{}
	_ servicemesh.HasDependencies  = &Service{}
	_ webRouter.HasRouteSlug       = &Service{}
	_ webRouter.IsRouteInitializer = &Service{}
	_ LoadsStringTables            = &Service{}
)

type Dependency = LoadsStringTables

type LoadsStringTables interface {
	GetSupportedLanguages() []string
	GetSupportedCharsets() []string
	GetTextByKey(key string) (string, error)
}
