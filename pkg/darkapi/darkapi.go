// Package darkapi provides access to the dark-magic API packages to be imported
// natively in Yaegi.
package darkapi

import "reflect"

// Symbols variable stores the map of dark-magic API symbols per package.
var Symbols = map[string]map[string]reflect.Value{}

func init() {
	Symbols["github.com/gravestench/dark-magic/pkg/darkapi"] = map[string]reflect.Value{
		"Symbols": reflect.ValueOf(Symbols),
	}
}

//go:generate ./gen.sh
