package luaManager

import (
	"fmt"
	"reflect"
	"unicode"

	lua "github.com/yuin/gopher-lua"
)

// createLuaTableFromStruct populates a Lua table with getters, Setters, and methods of a struct.
func createLuaTableFromStruct(L *lua.LState, v interface{}) *lua.LTable {
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
			nestedTable := createLuaTableFromStruct(L, field.Addr().Interface())
			L.SetField(t, fieldName, nestedTable)
		default:
			L.SetField(t, fieldName, L.NewFunction(func(L *lua.LState) int {
				switch kind {
				case reflect.Struct, reflect.Interface, reflect.Ptr:
					nestedTable := createLuaTableFromStruct(L, field.Addr().Interface())
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

func firstCharLowercase(s string) bool {
	if len(s) == 0 {
		return false
	}
	r := []rune(s)[0]
	return unicode.IsLower(r)
}
