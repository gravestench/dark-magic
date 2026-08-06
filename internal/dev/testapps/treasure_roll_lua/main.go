package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/gravestench/dark-magic/internal/content"
	"github.com/gravestench/dark-magic/internal/runtime/lua"
	lua "github.com/yuin/gopher-lua"
)

func main() {
	fileName := flag.String("file", "data/global/excel/treasureclassex.txt", "layered TreasureClass TSV path")
	className := flag.String("class", "", "treasure class to roll")
	seed := flag.Uint64("seed", 1, "deterministic seed")
	flag.Parse()
	if *className == "" {
		fmt.Fprintln(os.Stderr, "usage: treasure_roll_test_lua -class CLASS [-file path] [-seed N]")
		os.Exit(2)
	}
	contentFS, err := content.FromEnvironment()
	if err != nil {
		fatal(err)
	}
	runtime := modruntime.New()
	if err := runtime.RegisterModule(modruntime.LootModule(contentFS)); err != nil {
		fatal(err)
	}
	ctx := context.Background()
	if err := runtime.Start(ctx); err != nil {
		fatal(err)
	}
	defer runtime.Stop(ctx)
	var output any
	if err := runtime.Run(ctx, func(state *lua.LState) error {
		if err := state.CallByParam(lua.P{Fn: state.GetGlobal("require"), NRet: 1, Protect: true}, lua.LString("dm.loot/v1")); err != nil {
			return err
		}
		lootModule := state.Get(-1).(*lua.LTable)
		state.Pop(1)
		roll := lootModule.RawGetString("roll_tsv")
		if err := state.CallByParam(lua.P{Fn: roll, NRet: 2, Protect: true}, lua.LString(*fileName), lua.LString(*className), lua.LNumber(*seed)); err != nil {
			return err
		}
		errorValue := state.Get(-1)
		result := state.Get(-2)
		state.Pop(2)
		if result == lua.LNil {
			return fmt.Errorf("%s", errorValue)
		}
		converted, err := luaValue(result)
		output = converted
		return err
	}); err != nil {
		fatal(err)
	}
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(output); err != nil {
		fatal(err)
	}
}

func luaValue(value lua.LValue) (any, error) {
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
		if value.Len() > 0 {
			result := make([]any, value.Len())
			for index := range result {
				converted, err := luaValue(value.RawGetInt(index + 1))
				if err != nil {
					return nil, err
				}
				result[index] = converted
			}
			return result, nil
		}
		result := make(map[string]any)
		var conversionErr error
		value.ForEach(func(key, entry lua.LValue) {
			if conversionErr != nil {
				return
			}
			converted, err := luaValue(entry)
			if err != nil {
				conversionErr = err
				return
			}
			result[key.String()] = converted
		})
		return result, conversionErr
	default:
		return nil, fmt.Errorf("unsupported Lua value %s", value.Type())
	}
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, "treasure_roll_test_lua:", err)
	os.Exit(1)
}
