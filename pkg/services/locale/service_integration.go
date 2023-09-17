package locale

import (
	"github.com/gravestench/runtime"

	"github.com/gravestench/dark-magic/pkg/services/web_router"
)

var (
	_ runtime.Service               = &Service{}
	_ runtime.HasLogger             = &Service{}
	_ runtime.HasDependencies       = &Service{}
	_ web_router.HasRouteSlug       = &Service{}
	_ web_router.IsRouteInitializer = &Service{}
	_ LoadsStringTables             = &Service{}
)

type Dependency = LoadsStringTables

type LoadsStringTables interface {
	GetSupportedLanguages() []string
	GetSupportedCharsets() []string
	GetTextByKey(key string) (string, error)
}
