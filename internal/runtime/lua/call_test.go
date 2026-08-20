package modruntime

import (
	"testing"
	"testing/fstest"
)

// TestCallTransportsOnlySerializableValues protects the call transports only serializable values contract,
// including its observable ordering and failure behavior.
func TestCallTransportsOnlySerializableValues(t *testing.T) {
	runtime := New()
	if err := runtime.RegisterInstaller(ContentRequire(fstest.MapFS{
		"lua/example.lua": {
			Data: []byte(
				`return {choose=function(values) return {id=values[2].id,count=#values} end}`,
			),
		},
	}, "lua")); err != nil {
		t.Fatal(err)
	}

	if err := runtime.Start(t.Context()); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = runtime.Stop(t.Context()) }()

	value, err := Call(
		t.Context(),
		runtime,
		"example",
		"choose",
		[]any{map[string]any{"id": 1}, map[string]any{"id": 2}},
	)
	if err != nil {
		t.Fatal(err)
	}

	result := value.(map[string]any)
	if result["id"] != float64(2) || result["count"] != float64(2) {
		t.Fatalf("result = %#v", result)
	}
}
