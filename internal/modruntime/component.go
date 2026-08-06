package modruntime

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/fs"

	"github.com/gravestench/dark-magic/internal/host"
	lua "github.com/yuin/gopher-lua"
)

// Definition is a declarative Lua component loaded from content.
type Definition struct {
	ID        string
	DependsOn []string
	Source    string

	runtime *Runtime
	table   *lua.LTable
	start   *lua.LFunction
	stop    *lua.LFunction
	export  *lua.LFunction
	import_ *lua.LFunction
}

// LoadDefinition compiles source and requires it to return a component table.
func LoadDefinition(ctx context.Context, runtime *Runtime, source fs.FS, name string) (Definition, error) {
	data, err := fs.ReadFile(source, name)
	if err != nil {
		return Definition{}, fmt.Errorf("modruntime: read definition %q: %w", name, err)
	}
	var definition Definition
	err = runtime.Run(ctx, func(state *lua.LState) error {
		function, err := state.Load(bytes.NewReader(data), "@"+name)
		if err != nil {
			return fmt.Errorf("compile: %w", err)
		}
		state.SetFEnv(function, isolatedEnvironment(state))
		state.Push(function)
		if err := state.PCall(0, 1, nil); err != nil {
			return scriptError(name, "execute component", err)
		}
		value := state.Get(-1)
		state.Pop(1)
		table, ok := value.(*lua.LTable)
		if !ok {
			return fmt.Errorf("expected component table, got %s", value.Type())
		}
		idValue := table.RawGetString("id")
		id, ok := idValue.(lua.LString)
		if !ok || id == "" {
			return errors.New("component id must be a non-empty string")
		}
		definition = Definition{ID: string(id), Source: name, runtime: runtime, table: table}
		if dependencies := table.RawGetString("depends_on"); dependencies != lua.LNil {
			list, ok := dependencies.(*lua.LTable)
			if !ok {
				return errors.New("component depends_on must be an array")
			}
			var dependencyErr error
			list.ForEach(func(_, value lua.LValue) {
				if dependencyErr != nil {
					return
				}
				dependency, ok := value.(lua.LString)
				if !ok || dependency == "" {
					dependencyErr = errors.New("component dependencies must be non-empty strings")
					return
				}
				definition.DependsOn = append(definition.DependsOn, string(dependency))
			})
			if dependencyErr != nil {
				return dependencyErr
			}
		}
		if start := table.RawGetString("start"); start != lua.LNil {
			var ok bool
			definition.start, ok = start.(*lua.LFunction)
			if !ok {
				return errors.New("component start must be a function")
			}
		}
		if stop := table.RawGetString("stop"); stop != lua.LNil {
			var ok bool
			definition.stop, ok = stop.(*lua.LFunction)
			if !ok {
				return errors.New("component stop must be a function")
			}
		}
		if export := table.RawGetString("export_state"); export != lua.LNil {
			var ok bool
			definition.export, ok = export.(*lua.LFunction)
			if !ok {
				return errors.New("component export_state must be a function")
			}
		}
		if importState := table.RawGetString("import_state"); importState != lua.LNil {
			var ok bool
			definition.import_, ok = importState.(*lua.LFunction)
			if !ok {
				return errors.New("component import_state must be a function")
			}
		}
		return nil
	})
	if err != nil {
		return Definition{}, fmt.Errorf("modruntime: load definition %q: %w", name, err)
	}
	return definition, nil
}

// ReloadDefinition loads and transactionally replaces an existing script
// component. The current instance remains active if loading or replacement
// fails.
func ReloadDefinition(ctx context.Context, manager *host.Manager, runtime *Runtime, source fs.FS, name string) error {
	definition, err := LoadDefinition(ctx, runtime, source, name)
	if err != nil {
		return err
	}
	return manager.Replace(ctx, definition.Managed())
}

// Managed adapts a Lua definition to the shared runtime manager.
func (d Definition) Managed() host.ManagedDefinition {
	dependencies := append([]string(nil), d.DependsOn...)
	return host.ManagedDefinition{
		ID:        d.ID,
		DependsOn: dependencies,
		New: func(context.Context) (host.Component, error) {
			return &scriptComponent{definition: d, scope: &Scope{}}, nil
		},
	}
}

type scriptComponent struct {
	definition Definition
	scope      *Scope
}

func (c *scriptComponent) Start(ctx context.Context) error {
	if c.definition.start == nil {
		return nil
	}
	return c.definition.runtime.runScoped(ctx, c.scope, func(state *lua.LState) error {
		if err := state.CallByParam(lua.P{Fn: c.definition.start, NRet: 0, Protect: true}, c.definition.table); err != nil {
			return scriptError(c.definition.Source, "start component "+c.definition.ID, err)
		}
		return nil
	})
}

func (c *scriptComponent) Stop(ctx context.Context) error {
	var lifecycleErr error
	if c.definition.stop != nil {
		lifecycleErr = c.definition.runtime.runScoped(ctx, c.scope, func(state *lua.LState) error {
			if err := state.CallByParam(lua.P{Fn: c.definition.stop, NRet: 0, Protect: true}, c.definition.table); err != nil {
				return scriptError(c.definition.Source, "stop component "+c.definition.ID, err)
			}
			return nil
		})
	}
	return errors.Join(lifecycleErr, c.scope.Close())
}

func (c *scriptComponent) ExportState(ctx context.Context) (any, error) {
	if c.definition.export == nil {
		return nil, nil
	}
	var result any
	err := c.definition.runtime.runScoped(ctx, c.scope, func(state *lua.LState) error {
		if err := state.CallByParam(lua.P{Fn: c.definition.export, NRet: 1, Protect: true}, c.definition.table); err != nil {
			return scriptError(c.definition.Source, "export component "+c.definition.ID, err)
		}
		value := state.Get(-1)
		state.Pop(1)
		converted, err := luaToGo(value, 0)
		if err != nil {
			return fmt.Errorf("export Lua component %q state: %w", c.definition.ID, err)
		}
		result = converted
		return nil
	})
	return result, err
}

func (c *scriptComponent) ImportState(ctx context.Context, value any) error {
	if c.definition.import_ == nil || value == nil {
		return nil
	}
	return c.definition.runtime.runScoped(ctx, c.scope, func(state *lua.LState) error {
		converted, err := goToLua(state, value, 0)
		if err != nil {
			return fmt.Errorf("import Lua component %q state: %w", c.definition.ID, err)
		}
		if err := state.CallByParam(lua.P{Fn: c.definition.import_, NRet: 0, Protect: true}, c.definition.table, converted); err != nil {
			return scriptError(c.definition.Source, "import component "+c.definition.ID, err)
		}
		return nil
	})
}

func luaToGo(value lua.LValue, depth int) (any, error) {
	if depth > 64 {
		return nil, errors.New("state exceeds maximum nesting depth")
	}
	switch value := value.(type) {
	case *lua.LNilType:
		return nil, nil
	case lua.LBool:
		return bool(value), nil
	case lua.LNumber:
		return float64(value), nil
	case lua.LString:
		return string(value), nil
	case *lua.LTable:
		arrayLength := value.Len()
		array := make([]any, arrayLength)
		isArray := arrayLength > 0
		for index := 1; index <= arrayLength; index++ {
			converted, err := luaToGo(value.RawGetInt(index), depth+1)
			if err != nil {
				return nil, err
			}
			array[index-1] = converted
		}
		object := make(map[string]any)
		var conversionErr error
		value.ForEach(func(key, entry lua.LValue) {
			if conversionErr != nil {
				return
			}
			if number, ok := key.(lua.LNumber); ok && int(number) >= 1 && int(number) <= arrayLength {
				return
			}
			isArray = false
			name, ok := key.(lua.LString)
			if !ok {
				conversionErr = fmt.Errorf("state table key %s is not a string", key.Type())
				return
			}
			converted, err := luaToGo(entry, depth+1)
			if err != nil {
				conversionErr = err
				return
			}
			object[string(name)] = converted
		})
		if conversionErr != nil {
			return nil, conversionErr
		}
		if isArray {
			return array, nil
		}
		for index, entry := range array {
			object[fmt.Sprint(index+1)] = entry
		}
		return object, nil
	default:
		return nil, fmt.Errorf("state value type %s is not serializable", value.Type())
	}
}

func goToLua(state *lua.LState, value any, depth int) (lua.LValue, error) {
	if depth > 64 {
		return nil, errors.New("state exceeds maximum nesting depth")
	}
	switch value := value.(type) {
	case nil:
		return lua.LNil, nil
	case bool:
		return lua.LBool(value), nil
	case string:
		return lua.LString(value), nil
	case float64:
		return lua.LNumber(value), nil
	case []any:
		table := state.NewTable()
		for _, entry := range value {
			converted, err := goToLua(state, entry, depth+1)
			if err != nil {
				return nil, err
			}
			table.Append(converted)
		}
		return table, nil
	case map[string]any:
		table := state.NewTable()
		for key, entry := range value {
			converted, err := goToLua(state, entry, depth+1)
			if err != nil {
				return nil, err
			}
			table.RawSetString(key, converted)
		}
		return table, nil
	default:
		return nil, fmt.Errorf("Go state value type %T is not serializable", value)
	}
}
