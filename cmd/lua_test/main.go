package main

import (
	_ "embed"
	"fmt"
	"reflect"
	"unicode"

	lua "github.com/yuin/gopher-lua"
)

func main() {
	L := lua.NewState()
	defer L.Close()

	// this is using reflection to build up a lua table object which has all
	// setters/getters for exported fields, as well as any methods.
	table := PopulateTableWithStruct(L, &Body{})
	L.SetGlobal("example", table)

	if err := L.DoString(script); err != nil {
		fmt.Print("\n")
		defer fmt.Print("\n\n")
		panic(err)
	}
}

//go:embed script.lua
var script string

// PopulateTableWithStruct populates a Lua table with getters, Setters, and methods of a struct.
func PopulateTableWithStruct(L *lua.LState, v interface{}) *lua.LTable {
	t := L.NewTable()
	val := reflect.ValueOf(v)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	typ := val.Type()

	// Populate getters and Setters for fields
	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldName := typ.Field(i).Name
		kind := field.Kind()

		if firstCharLowercase(fieldName) {
			continue
		}

		// Getter
		switch kind {
		case reflect.Struct:
			// getter for a struct needs to be a table, not a function
			nestedTable := PopulateTableWithStruct(L, field.Addr().Interface())
			L.SetField(t, fieldName, nestedTable)
		default:
			L.SetField(t, fieldName, L.NewFunction(func(L *lua.LState) int {
				switch kind {
				case reflect.Struct, reflect.Interface, reflect.Ptr:
					nestedTable := PopulateTableWithStruct(L, field.Addr().Interface())
					L.Push(nestedTable)
				default:
					L.Push(lua.LString(fmt.Sprint(field.Interface())))
				}

				return 1
			}))
		}

		switch kind {
		case reflect.String, reflect.Int, reflect.Float64:
		default:
			continue
		}

		// Setter
		L.SetField(t, "Set"+fieldName, L.NewFunction(func(L *lua.LState) int {
			if field.CanSet() && kind != reflect.Struct {
				switch kind {
				case reflect.String:
					field.SetString(L.CheckString(1))
				case reflect.Int:
					field.SetInt(int64(L.CheckInt(1)))
				case reflect.Float64:
					field.SetFloat(float64(L.CheckNumber(1)))
				case reflect.Struct:
					// Handling for nested structs could be added here
					// Currently not allowing direct Set on nested structs
					// Add other types as needed
				}
			}
			return 0
		}))
	}

	// Reflect on the pointer type to get methods
	ptrVal := reflect.ValueOf(v)
	ptrType := reflect.PtrTo(typ)
	numMethods := ptrType.NumMethod()

	// Populate methods
	for i := 0; i < numMethods; i++ {
		method := ptrType.Method(i)
		methodName := method.Name

		if firstCharLowercase(methodName) {
			continue
		}

		L.SetField(t, methodName, L.NewFunction(func(L *lua.LState) int {
			method := method.Func
			numIn := method.Type().NumIn()
			args := make([]reflect.Value, numIn)
			args[0] = ptrVal

			for j := 1; j < numIn; j++ {
				switch method.Type().In(j).Kind() {
				case reflect.String:
					args[j] = reflect.ValueOf(L.CheckString(j))
				case reflect.Int:
					args[j] = reflect.ValueOf(L.CheckInt(j))
				case reflect.Float64:
					args[j] = reflect.ValueOf(L.CheckNumber(j))
				case reflect.Struct:
					// Handle custom struct arguments if needed
					// Assuming structs are passed as tables
					// args[j] = reflect.ValueOf(ConvertLuaTableToStruct(L, j, method.Type().In(j).Elem()))
				// Add other types as needed
				default:
					args[j] = reflect.Zero(method.Type().In(j))
				}
			}

			results := method.Call(args)
			for _, result := range results {
				L.Push(lua.LString(fmt.Sprint(result.Interface())))
			}
			return len(results)
		}))
	}

	return t
}

// Body struct with fields and methods
type Body struct {
	Position
	Acceleration
	thisIsPrivate int
}

type Position struct {
	Vec3
}

type Acceleration struct {
	Vec3
}

type Vec2 struct {
	Public string
	p1, p2 float64
}

func (v *Vec2) XY() (x, y float64) { return v.p1, v.p2 }

type Vec3 struct {
	Vec2
	p3 float64
}

func (*Vec3) private() {
	return
}

func (v *Vec3) XYZ() (x, y, z float64) { return v.p1, v.p2, v.p3 }

func (e *Body) private() {
	return
}

func (e *Body) Method1(arg string) string {
	return "Method1 called with: " + arg
}

func (e *Body) AsString() string {
	return fmt.Sprintf("%+v", e)
}

// IsFirstCharLowercase checks if the first character of a string is lowercase.
func firstCharLowercase(s string) bool {
	if len(s) == 0 {
		return false
	}
	r := []rune(s)[0]
	return unicode.IsLower(r)
}
