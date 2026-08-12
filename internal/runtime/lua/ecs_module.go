package modruntime

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"

	"github.com/gravestench/akara"
	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
	lua "github.com/yuin/gopher-lua"
)

const (
	ecsEntityType    = "engine.ecs.entity/v1"
	ecsComponentType = "engine.ecs.component/v1"
	ecsCommandsType  = "engine.ecs.commands/v1"
)

type ecsAccess struct {
	read, write map[string]struct{}
}

// ECSCapability adapts one authoritative engine into checked Lua values. It
// tracks the currently executing system's declared access so scripts cannot use
// a component reference to bypass read/write policy.
type ECSCapability struct {
	runtime *Runtime
	engine  *gameecs.Engine
	active  *ecsAccess // Lua-owner goroutine only
}

type ownedECSEntity struct {
	engine *gameecs.Engine
	entity akara.Entity
}
type ownedECSComponent struct {
	capability *ECSCapability
	name       string
	component  *akara.DynamicComponent
}
type ownedECSCommands struct {
	capability *ECSCapability
	commands   *gameecs.StructuralCommands
}

// NewECSCapability creates the adapter; the runtime and engine retain ownership.
func NewECSCapability(runtime *Runtime, engine *gameecs.Engine) *ECSCapability {
	return &ECSCapability{runtime: runtime, engine: engine}
}

// Module returns the versioned engine.ecs/v1 registration. Resource-producing
// operations attach their cleanup to the active Lua scope.
func (capability *ECSCapability) Module() Module {
	return Module{Name: "engine.ecs/v1", Help: documentedModule("Define runtime ECS components and deterministic scoped systems.", map[string]CommandHelp{
		"component": commandHelp("engine.ecs.component(definition)", "Register or migrate a named runtime component schema."),
		"create":    commandHelp("engine.ecs.create([components])", "Create an entity and optionally attach named components."),
		"get":       commandHelp("engine.ecs.get(entity, component)", "Return a checked component reference or nil."),
		"query":     commandHelp("engine.ecs.query(filter)", "Return an ordered snapshot of entities matching all, any, and none component filters."),
		"set":       commandHelp("engine.ecs.set(entity, component, values)", "Add or replace a component after schema validation."),
		"remove":    commandHelp("engine.ecs.remove(entity, component)", "Remove a component immediately outside system iteration."),
		"destroy":   commandHelp("engine.ecs.destroy(entity)", "Destroy an entity immediately outside system iteration."),
		"system":    commandHelp("engine.ecs.system(definition)", "Register a scope-owned ordered simulation system."),
	}), Loader: func(state *lua.LState) int {
		registerECSTypes(state)
		module := state.SetFuncs(state.NewTable(), map[string]lua.LGFunction{
			"component": capability.loadComponent,
			"create":    capability.createEntity,
			"get":       capability.getComponent,
			"query":     capability.queryEntities,
			"set":       capability.setComponent,
			"remove":    capability.removeComponent,
			"destroy":   capability.destroyEntity,
			"system":    capability.registerSystem,
		})
		module.RawSetString("api", lua.LNumber(1))
		state.Push(module)
		return 1
	}}
}

func (capability *ECSCapability) queryEntities(state *lua.LState) int {
	if capability.active != nil {
		state.RaiseError("ecs.query cannot be called inside a system; use the system query entities")
	}
	query := state.CheckTable(1)
	all, err := capability.stores(tableStringList(state, query.RawGetString("all")))
	if err != nil {
		state.RaiseError("%v", err)
	}
	any, err := capability.stores(tableStringList(state, query.RawGetString("any")))
	if err != nil {
		state.RaiseError("%v", err)
	}
	none, err := capability.stores(tableStringList(state, query.RawGetString("none")))
	if err != nil {
		state.RaiseError("%v", err)
	}
	options := make([]akara.FilterOption, 0, 3)
	if len(all) > 0 {
		options = append(options, akara.All(componentTypes(all)...))
	}
	if len(any) > 0 {
		options = append(options, akara.Any(componentTypes(any)...))
	}
	if len(none) > 0 {
		options = append(options, akara.None(componentTypes(none)...))
	}
	subscription, err := capability.engine.World().Subscribe(options...)
	if err != nil {
		state.RaiseError("%v", err)
	}
	entities := subscription.Entities()
	subscription.Close()
	result := state.NewTable()
	for _, entity := range entities {
		result.Append(capability.entityValue(state, entity))
	}
	state.Push(result)
	return 1
}

func registerECSTypes(state *lua.LState) {
	entityMeta := state.NewTypeMetatable(ecsEntityType)
	state.SetField(entityMeta, "__index", state.SetFuncs(state.NewTable(), map[string]lua.LGFunction{
		"id": func(state *lua.LState) int { state.Push(lua.LNumber(checkECSEntity(state, 1).entity)); return 1 },
		"exists": func(state *lua.LState) int {
			entity := checkECSEntity(state, 1)
			state.Push(lua.LBool(entity.engine.World().EntityExists(entity.entity)))
			return 1
		},
	}))
	componentMeta := state.NewTypeMetatable(ecsComponentType)
	state.SetField(componentMeta, "__index", state.SetFuncs(state.NewTable(), map[string]lua.LGFunction{
		"get": func(state *lua.LState) int {
			component := checkECSComponent(state, 1)
			value, err := component.read(state.CheckString(2))
			if err != nil {
				state.RaiseError("%v", err)
			}
			state.Push(dynamicToLua(state, component.capability.engine, value))
			return 1
		},
		"set": func(state *lua.LState) int {
			component := checkECSComponent(state, 1)
			if err := component.write(state.CheckString(2), state.Get(3)); err != nil {
				state.RaiseError("%v", err)
			}
			return 0
		},
		"snapshot": func(state *lua.LState) int {
			component := checkECSComponent(state, 1)
			values, err := component.snapshot()
			if err != nil {
				state.RaiseError("%v", err)
			}
			state.Push(dynamicMapToLua(state, component.capability.engine, values))
			return 1
		},
	}))
	commandsMeta := state.NewTypeMetatable(ecsCommandsType)
	state.SetField(commandsMeta, "__index", state.SetFuncs(state.NewTable(), map[string]lua.LGFunction{
		"create":  commandCreate,
		"set":     commandSet,
		"remove":  commandRemove,
		"destroy": commandDestroy,
	}))
}

func (capability *ECSCapability) loadComponent(state *lua.LState) int {
	definition := state.CheckTable(1)
	name := requiredTableString(state, definition, "name")
	version := uint32(tableInteger(state, definition, "version", 1))
	fieldsTable, ok := definition.RawGetString("fields").(*lua.LTable)
	if !ok {
		state.ArgError(1, "fields must be an array")
	}
	fields := make([]akara.Field, 0, fieldsTable.Len())
	fieldsTable.ForEach(func(_, value lua.LValue) {
		fieldTable, ok := value.(*lua.LTable)
		if !ok {
			state.RaiseError("component fields must be tables")
		}
		kindName := requiredTableString(state, fieldTable, "type")
		kind, ok := luaFieldKinds[kindName]
		if !ok {
			state.RaiseError("unknown ECS field type %q", kindName)
		}
		field := akara.Field{Name: requiredTableString(state, fieldTable, "name"), Kind: kind}
		if defaultValue := fieldTable.RawGetString("default"); defaultValue != lua.LNil {
			field.Default = luaToDynamic(state, kind, defaultValue)
		}
		fields = append(fields, field)
	})
	schema := akara.Schema{Name: name, Version: version, Fields: fields}
	if existing, found := akara.GetDynamicStore(capability.engine.World(), name); found {
		current := existing.Schema()
		if current.Version == version {
			if _, err := akara.RegisterSchema(capability.engine.World(), schema); err != nil {
				state.RaiseError("%v", err)
			}
		} else {
			migrate, ok := definition.RawGetString("migrate").(*lua.LFunction)
			if !ok {
				state.RaiseError("component %q version %d requires migrate(previous, entity)", name, version)
			}
			err := existing.Migrate(schema, func(entity akara.Entity, previous map[string]any) (map[string]any, error) {
				if err := state.CallByParam(lua.P{Fn: migrate, NRet: 1, Protect: true}, dynamicMapToLua(state, capability.engine, previous), capability.entityValue(state, entity)); err != nil {
					return nil, err
				}
				value := state.Get(-1)
				state.Pop(1)
				return luaDynamicMapForSchema(value, schema)
			})
			if err != nil {
				state.RaiseError("%v", err)
			}
		}
	} else if _, err := akara.RegisterSchema(capability.engine.World(), schema); err != nil {
		state.RaiseError("%v", err)
	}
	state.Push(lua.LString(name))
	return 1
}

var luaFieldKinds = map[string]akara.FieldKind{
	"bool": akara.FieldBool, "i64": akara.FieldInt64, "u64": akara.FieldUint64,
	"f64": akara.FieldFloat64, "string": akara.FieldString, "entity": akara.FieldEntity,
}

func (capability *ECSCapability) createEntity(state *lua.LState) int {
	if capability.active != nil {
		state.RaiseError("entity creation inside a system requires commands:create")
	}
	entity, err := capability.engine.World().CreateEntity()
	if err != nil {
		state.RaiseError("%v", err)
	}
	if values, ok := state.Get(1).(*lua.LTable); ok {
		var setErr error
		values.ForEach(func(key, value lua.LValue) {
			if setErr != nil {
				return
			}
			name, ok := key.(lua.LString)
			if !ok {
				setErr = errors.New("component names must be strings")
				return
			}
			_, setErr = capability.setDynamic(entity, string(name), value)
		})
		if setErr != nil {
			capability.engine.World().DestroyEntity(entity)
			state.RaiseError("%v", setErr)
		}
	}
	state.Push(capability.entityValue(state, entity))
	return 1
}

func (capability *ECSCapability) getComponent(state *lua.LState) int {
	entity, name := capability.checkEntity(state, 1), state.CheckString(2)
	if err := capability.requireRead(name); err != nil {
		state.RaiseError("%v", err)
	}
	store, found := akara.GetDynamicStore(capability.engine.World(), name)
	if !found {
		state.RaiseError("unknown component %q", name)
	}
	component, found := store.Get(entity.entity)
	if !found {
		state.Push(lua.LNil)
		return 1
	}
	state.Push(capability.componentValue(state, name, component))
	return 1
}

func (capability *ECSCapability) setComponent(state *lua.LState) int {
	entity, name := capability.checkEntity(state, 1), state.CheckString(2)
	if err := capability.requireWrite(name); err != nil {
		state.RaiseError("%v", err)
	}
	if capability.active != nil {
		state.RaiseError("component replacement inside a system requires commands:set; mutate existing fields through the component reference")
	}
	component, err := capability.setDynamic(entity.entity, name, state.Get(3))
	if err != nil {
		state.RaiseError("%v", err)
	}
	state.Push(capability.componentValue(state, name, component))
	return 1
}

func (capability *ECSCapability) removeComponent(state *lua.LState) int {
	entity, name := capability.checkEntity(state, 1), state.CheckString(2)
	if capability.active != nil {
		state.RaiseError("structural changes inside a system require commands:remove")
	}
	store, found := akara.GetDynamicStore(capability.engine.World(), name)
	if !found {
		state.RaiseError("unknown component %q", name)
	}
	state.Push(lua.LBool(store.Remove(entity.entity)))
	return 1
}

func (capability *ECSCapability) destroyEntity(state *lua.LState) int {
	if capability.active != nil {
		state.RaiseError("structural changes inside a system require commands:destroy")
	}
	state.Push(lua.LBool(capability.engine.World().DestroyEntity(capability.checkEntity(state, 1).entity)))
	return 1
}

func (capability *ECSCapability) registerSystem(state *lua.LState) int {
	scope, err := capability.runtime.requireActiveScope()
	if err != nil {
		state.RaiseError("%v", err)
	}
	table := state.CheckTable(1)
	id := requiredTableString(state, table, "id")
	phase := gameecs.Phase(requiredTableString(state, table, "phase"))
	update, ok := table.RawGetString("update").(*lua.LFunction)
	if !ok {
		state.ArgError(1, "update must be a function")
	}
	readNames := tableStringList(state, table.RawGetString("read"))
	writeNames := tableStringList(state, table.RawGetString("write"))
	read, err := capability.stores(readNames)
	if err != nil {
		state.RaiseError("%v", err)
	}
	write, err := capability.stores(writeNames)
	if err != nil {
		state.RaiseError("%v", err)
	}
	query, _ := table.RawGetString("query").(*lua.LTable)
	allNames, anyNames, noneNames := []string{}, []string{}, []string{}
	if query != nil {
		allNames = tableStringList(state, query.RawGetString("all"))
		anyNames = tableStringList(state, query.RawGetString("any"))
		noneNames = tableStringList(state, query.RawGetString("none"))
	}
	all, err := capability.stores(allNames)
	if err != nil {
		state.RaiseError("%v", err)
	}
	any, err := capability.stores(anyNames)
	if err != nil {
		state.RaiseError("%v", err)
	}
	none, err := capability.stores(noneNames)
	if err != nil {
		state.RaiseError("%v", err)
	}
	access := &ecsAccess{read: nameSet(readNames), write: nameSet(writeNames)}
	definition := gameecs.Definition{ID: id, Phase: phase, After: tableStringList(state, table.RawGetString("after")), Before: tableStringList(state, table.RawGetString("before")), All: componentTypes(all), Any: componentTypes(any), None: componentTypes(none), Read: componentTypes(read), Write: componentTypes(write)}
	definition.Update = func(ctx gameecs.Context, entities []akara.Entity, commands *gameecs.StructuralCommands) error {
		return capability.runtime.runScoped(context.Background(), scope, func(state *lua.LState) error {
			previous := capability.active
			capability.active = access
			defer func() { capability.active = previous }()
			contextTable := state.NewTable()
			contextTable.RawSetString("tick", lua.LNumber(ctx.Tick))
			contextTable.RawSetString("delta_seconds", lua.LNumber(ctx.Delta.Seconds()))
			entityTable := state.NewTable()
			for _, entity := range entities {
				entityTable.Append(capability.entityValue(state, entity))
			}
			commandsValue := state.NewUserData()
			commandsValue.Value = &ownedECSCommands{capability: capability, commands: commands}
			state.SetMetatable(commandsValue, state.GetTypeMetatable(ecsCommandsType))
			if err := state.CallByParam(lua.P{Fn: update, NRet: 0, Protect: true}, contextTable, entityTable, commandsValue); err != nil {
				return err
			}
			return nil
		})
	}
	if err := capability.engine.Register(definition); err != nil {
		state.RaiseError("%v", err)
	}
	if err := scope.Add(func() error {
		if !capability.engine.Unregister(id) {
			return fmt.Errorf("game ecs: unregister %q", id)
		}
		return nil
	}); err != nil {
		capability.engine.Unregister(id)
		state.RaiseError("%v", err)
	}
	state.Push(lua.LString(id))
	return 1
}

func (capability *ECSCapability) stores(names []string) ([]*akara.DynamicStore, error) {
	result := make([]*akara.DynamicStore, len(names))
	for index, name := range names {
		store, found := akara.GetDynamicStore(capability.engine.World(), name)
		if !found {
			return nil, fmt.Errorf("unknown component %q", name)
		}
		result[index] = store
	}
	return result, nil
}
func componentTypes(stores []*akara.DynamicStore) []akara.ComponentType {
	result := make([]akara.ComponentType, len(stores))
	for i, store := range stores {
		result[i] = store
	}
	return result
}
func nameSet(names []string) map[string]struct{} {
	result := make(map[string]struct{}, len(names))
	for _, name := range names {
		result[name] = struct{}{}
	}
	return result
}

func (capability *ECSCapability) requireRead(name string) error {
	if capability.active == nil {
		return nil
	}
	if _, ok := capability.active.read[name]; ok {
		return nil
	}
	if _, ok := capability.active.write[name]; ok {
		return nil
	}
	return fmt.Errorf("system did not declare read access to %q", name)
}
func (capability *ECSCapability) requireWrite(name string) error {
	if capability.active == nil {
		return nil
	}
	if _, ok := capability.active.write[name]; ok {
		return nil
	}
	return fmt.Errorf("system did not declare write access to %q", name)
}

func (capability *ECSCapability) setDynamic(entity akara.Entity, name string, value lua.LValue) (*akara.DynamicComponent, error) {
	store, found := akara.GetDynamicStore(capability.engine.World(), name)
	if !found {
		return nil, fmt.Errorf("unknown component %q", name)
	}
	values, err := luaDynamicMapForSchema(value, store.Schema())
	if err != nil {
		return nil, err
	}
	return store.Set(entity, values)
}

func (capability *ECSCapability) entityValue(state *lua.LState, entity akara.Entity) *lua.LUserData {
	value := state.NewUserData()
	value.Value = &ownedECSEntity{engine: capability.engine, entity: entity}
	state.SetMetatable(value, state.GetTypeMetatable(ecsEntityType))
	return value
}
func (capability *ECSCapability) componentValue(state *lua.LState, name string, component *akara.DynamicComponent) *lua.LUserData {
	value := state.NewUserData()
	value.Value = &ownedECSComponent{capability: capability, name: name, component: component}
	state.SetMetatable(value, state.GetTypeMetatable(ecsComponentType))
	return value
}

func (component *ownedECSComponent) read(name string) (any, error) {
	if err := component.capability.requireRead(component.name); err != nil {
		return nil, err
	}
	return component.component.Get(name)
}
func (component *ownedECSComponent) write(name string, value lua.LValue) error {
	if err := component.capability.requireWrite(component.name); err != nil {
		return err
	}
	store, _ := akara.GetDynamicStore(component.capability.engine.World(), component.name)
	field, ok := schemaField(store.Schema(), name)
	if !ok {
		return fmt.Errorf("unknown field %q", name)
	}
	return component.component.Set(name, luaToDynamic(nil, field.Kind, value))
}
func (component *ownedECSComponent) snapshot() (map[string]any, error) {
	if err := component.capability.requireRead(component.name); err != nil {
		return nil, err
	}
	return component.component.Snapshot()
}

func commandSet(state *lua.LState) int {
	owned := checkECSCommands(state, 1)
	entity := owned.capability.checkEntity(state, 2)
	name := state.CheckString(3)
	if err := owned.capability.requireWrite(name); err != nil {
		state.RaiseError("%v", err)
	}
	store, found := akara.GetDynamicStore(owned.capability.engine.World(), name)
	if !found {
		state.RaiseError("unknown component %q", name)
	}
	values, err := luaDynamicMapForSchema(state.Get(4), store.Schema())
	if err != nil {
		state.RaiseError("%v", err)
	}
	owned.commands.AddDynamic(store, entity.entity, values)
	return 0
}

func commandCreate(state *lua.LState) int {
	owned := checkECSCommands(state, 1)
	componentsTable := state.CheckTable(2)
	components := make(map[*akara.DynamicStore]map[string]any, componentsTable.Len())
	componentsTable.ForEach(func(key, value lua.LValue) {
		name, ok := key.(lua.LString)
		if !ok {
			state.RaiseError("component names must be strings")
		}
		if err := owned.capability.requireWrite(string(name)); err != nil {
			state.RaiseError("%v", err)
		}
		store, found := akara.GetDynamicStore(owned.capability.engine.World(), string(name))
		if !found {
			state.RaiseError("unknown component %q", name)
		}
		values, err := luaDynamicMapForSchema(value, store.Schema())
		if err != nil {
			state.RaiseError("%v", err)
		}
		components[store] = values
	})
	owned.commands.CreateDynamic(owned.capability.engine.World(), components)
	return 0
}
func commandRemove(state *lua.LState) int {
	owned := checkECSCommands(state, 1)
	entity := owned.capability.checkEntity(state, 2)
	name := state.CheckString(3)
	if err := owned.capability.requireWrite(name); err != nil {
		state.RaiseError("%v", err)
	}
	store, found := akara.GetDynamicStore(owned.capability.engine.World(), name)
	if !found {
		state.RaiseError("unknown component %q", name)
	}
	owned.commands.Remove(store, entity.entity)
	return 0
}
func commandDestroy(state *lua.LState) int {
	owned := checkECSCommands(state, 1)
	entity := owned.capability.checkEntity(state, 2)
	owned.commands.Destroy(owned.capability.engine.World(), entity.entity)
	return 0
}

func checkECSEntity(state *lua.LState, index int) *ownedECSEntity {
	value := state.CheckUserData(index)
	entity, ok := value.Value.(*ownedECSEntity)
	if !ok {
		state.ArgError(index, "engine.ecs/v1 entity expected")
	}
	return entity
}

func (capability *ECSCapability) checkEntity(state *lua.LState, index int) *ownedECSEntity {
	entity := checkECSEntity(state, index)
	if entity.engine != capability.engine {
		state.ArgError(index, "entity belongs to a different engine.ecs/v1 world")
	}
	return entity
}
func checkECSComponent(state *lua.LState, index int) *ownedECSComponent {
	value := state.CheckUserData(index)
	component, ok := value.Value.(*ownedECSComponent)
	if !ok {
		state.ArgError(index, "engine.ecs/v1 component expected")
	}
	return component
}
func checkECSCommands(state *lua.LState, index int) *ownedECSCommands {
	value := state.CheckUserData(index)
	commands, ok := value.Value.(*ownedECSCommands)
	if !ok {
		state.ArgError(index, "engine.ecs/v1 commands expected")
	}
	return commands
}

func requiredTableString(state *lua.LState, table *lua.LTable, name string) string {
	value, ok := table.RawGetString(name).(lua.LString)
	if !ok || strings.TrimSpace(string(value)) == "" {
		state.RaiseError("%s must be a non-empty string", name)
	}
	return string(value)
}
func tableInteger(state *lua.LState, table *lua.LTable, name string, fallback int) int {
	value := table.RawGetString(name)
	if value == lua.LNil {
		return fallback
	}
	number, ok := value.(lua.LNumber)
	if !ok || math.Trunc(float64(number)) != float64(number) {
		state.RaiseError("%s must be an integer", name)
	}
	return int(number)
}
func tableStringList(state *lua.LState, value lua.LValue) []string {
	if value == lua.LNil {
		return nil
	}
	table, ok := value.(*lua.LTable)
	if !ok {
		state.RaiseError("expected string array")
	}
	result := make([]string, 0, table.Len())
	table.ForEach(func(_, value lua.LValue) {
		text, ok := value.(lua.LString)
		if !ok || text == "" {
			state.RaiseError("expected non-empty string array")
		}
		result = append(result, string(text))
	})
	return result
}
func schemaField(schema akara.Schema, name string) (akara.Field, bool) {
	for _, field := range schema.Fields {
		if field.Name == name {
			return field, true
		}
	}
	return akara.Field{}, false
}

func luaDynamicMapForSchema(value lua.LValue, schema akara.Schema) (map[string]any, error) {
	if value == lua.LNil {
		return nil, nil
	}
	table, ok := value.(*lua.LTable)
	if !ok {
		return nil, errors.New("component values must be a table")
	}
	result := make(map[string]any)
	var conversionErr error
	table.ForEach(func(key, value lua.LValue) {
		if conversionErr != nil {
			return
		}
		name, ok := key.(lua.LString)
		if !ok {
			conversionErr = errors.New("component field names must be strings")
			return
		}
		field, found := schemaField(schema, string(name))
		if !found {
			conversionErr = fmt.Errorf("unknown field %q", name)
			return
		}
		result[string(name)] = luaToDynamic(nil, field.Kind, value)
	})
	return result, conversionErr
}
func luaToDynamic(_ *lua.LState, kind akara.FieldKind, value lua.LValue) any {
	switch kind {
	case akara.FieldBool:
		if typed, ok := value.(lua.LBool); ok {
			return bool(typed)
		}
	case akara.FieldInt64:
		if typed, ok := value.(lua.LNumber); ok && math.Trunc(float64(typed)) == float64(typed) {
			return int64(typed)
		}
	case akara.FieldUint64:
		if typed, ok := value.(lua.LNumber); ok && typed >= 0 && math.Trunc(float64(typed)) == float64(typed) {
			return uint64(typed)
		}
	case akara.FieldFloat64:
		if typed, ok := value.(lua.LNumber); ok {
			return float64(typed)
		}
	case akara.FieldString:
		if typed, ok := value.(lua.LString); ok {
			return string(typed)
		}
	case akara.FieldEntity:
		if typed, ok := value.(*lua.LUserData); ok {
			if entity, ok := typed.Value.(*ownedECSEntity); ok {
				return entity.entity
			}
		}
	}
	return value
}
func dynamicToLua(state *lua.LState, engine *gameecs.Engine, value any) lua.LValue {
	switch value := value.(type) {
	case bool:
		return lua.LBool(value)
	case int64:
		return lua.LNumber(value)
	case uint64:
		return lua.LNumber(value)
	case float64:
		return lua.LNumber(value)
	case string:
		return lua.LString(value)
	case akara.Entity:
		entity := state.NewUserData()
		entity.Value = &ownedECSEntity{engine: engine, entity: value}
		state.SetMetatable(entity, state.GetTypeMetatable(ecsEntityType))
		return entity
	default:
		return lua.LNil
	}
}
func dynamicMapToLua(state *lua.LState, engine *gameecs.Engine, values map[string]any) *lua.LTable {
	table := state.NewTable()
	for name, value := range values {
		table.RawSetString(name, dynamicToLua(state, engine, value))
	}
	return table
}
