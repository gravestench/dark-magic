package luashell

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	glua "github.com/yuin/gopher-lua"
)

// installShellOutput routes print and call-frame diagnostics into the transcript
// instead of process stdout/stderr, preserving output ordering with return values.
func installShellOutput(state *glua.LState, environment *glua.LTable, output *[]string) {
	environment.RawSetString("print", shellPrintFunction(state, output))

	registerFunction := shellRegisterFunction(state, output)

	// GopherLua's private _printregs writes directly to process stderr. Shadow
	// it inside shell scopes and provide a friendly alias that captures the same
	// class of debugging information in the shell transcript.
	environment.RawSetString("_printregs", registerFunction)
	environment.RawSetString("printregs", registerFunction)
}

// shellPrintFunction preserves Lua's tab-separated print conversion while
// appending one transcript line per call.
func shellPrintFunction(state *glua.LState, output *[]string) *glua.LFunction {
	return state.NewFunction(func(current *glua.LState) int {
		values := make([]string, current.GetTop())
		for index := 1; index <= current.GetTop(); index++ {
			values[index-1] = current.ToStringMeta(current.Get(index)).String()
		}

		*output = append(*output, strings.Join(values, "\t"))

		return 0
	})
}

// shellRegisterFunction formats call frames and locals without exposing them on
// process stderr, which keeps remote and embedded shells presentation-neutral.
func shellRegisterFunction(state *glua.LState, output *[]string) *glua.LFunction {
	return state.NewFunction(func(current *glua.LState) int {
		lines := []string{"Lua call frames:"}

		for level := 1; ; level++ {
			frame, ok := current.GetStack(level)
			if !ok {
				break
			}

			_, _ = current.GetInfo("nSl", frame, glua.LNil)

			location := frame.Source
			if frame.CurrentLine >= 0 {
				location += fmt.Sprintf(":%d", frame.CurrentLine)
			}

			lines = append(lines, fmt.Sprintf("  [%d] %s %s", level, frame.Name, location))
			lines = append(lines, shellFrameLocals(current, frame)...)
		}

		*output = append(*output, strings.Join(lines, "\n"))

		return 0
	})
}

// shellFrameLocals reads locals in Lua stack order so diagnostics retain the
// same positional meaning as GopherLua's frame inspection API.
func shellFrameLocals(state *glua.LState, frame *glua.Debug) []string {
	var lines []string

	for local := 1; ; local++ {
		name, value := state.GetLocal(frame, local)
		if name == "" {
			break
		}

		lines = append(lines, fmt.Sprintf("      %s = %s", name, formatValue(value, 0)))
	}

	return lines
}

// formatValue renders one Lua value with cycle detection and bounded table depth
// so arbitrary session state cannot recurse forever or traverse tables without limit.
func formatValue(value glua.LValue, depth int) string {
	return formatValueSeen(value, depth, make(map[*glua.LTable]bool))
}

// formatValueSeen recursively renders tables while sharing the active recursion
// set; removing tables on return allows repeated non-cyclic references to render.
func formatValueSeen(value glua.LValue, depth int, seen map[*glua.LTable]bool) string {
	switch typed := value.(type) {
	case *glua.LTable:
		return formatTable(typed, depth, seen)
	case glua.LString:
		if strings.ContainsAny(string(typed), "\r\n\t") {
			return string(typed)
		}

		return strconv.Quote(string(typed))
	case *glua.LFunction:
		return "<function>"
	case *glua.LUserData:
		return "<userdata>"
	default:
		return value.String()
	}
}

// formatTable sorts keys for deterministic output, marks cycles, and caps both
// recursion depth and member count to protect interactive rendering.
func formatTable(table *glua.LTable, depth int, seen map[*glua.LTable]bool) string {
	if seen[table] {
		return "<cycle>"
	}

	if depth >= 4 {
		return "{…}"
	}

	seen[table] = true
	defer delete(seen, table)

	parts := make([]string, 0, 16)
	truncated := false

	table.ForEach(func(key, item glua.LValue) {
		if len(parts) < 64 {
			parts = append(parts, key.String()+" = "+formatValueSeen(item, depth+1, seen))
		} else {
			truncated = true
		}
	})

	sort.Strings(parts)

	if truncated {
		parts = append(parts, "…")
	}

	if len(parts) == 0 {
		return "{}"
	}

	indent := strings.Repeat("  ", depth+1)
	closeIndent := strings.Repeat("  ", depth)

	return "{\n" + indent + strings.Join(parts, ",\n"+indent) + "\n" + closeIndent + "}"
}
