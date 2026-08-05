package luaManager

import (
	"errors"
	"io"
	"log/slog"
	"sync"
	"testing"

	"github.com/yuin/gopher-lua"
)

func newTestService() *Service {
	service := &Service{}
	service.SetLogger(slog.New(slog.NewTextHandler(io.Discard, nil)))
	service.RebuildState()
	return service
}

func TestShutdownClosesStateUntilExplicitRebuild(t *testing.T) {
	service := newTestService()
	service.OnShutdown()

	err := service.WithState(func(*lua.LState) error {
		t.Fatal("callback must not run after shutdown")
		return nil
	})
	if !errors.Is(err, ErrStateClosed) {
		t.Fatalf("WithState error = %v, want %v", err, ErrStateClosed)
	}

	service.RebuildState()
	defer service.OnShutdown()
	if err := service.WithState(func(*lua.LState) error { return nil }); err != nil {
		t.Fatalf("using explicitly rebuilt state: %v", err)
	}
}

func TestRebuildStateCreatesAPI(t *testing.T) {
	service := newTestService()
	defer service.OnShutdown()

	if !service.GlobalsExist("api") {
		t.Fatal("expected the root API table to exist")
	}
}

func TestWithStateSerializesConcurrentAccess(t *testing.T) {
	service := newTestService()
	defer service.OnShutdown()

	const workers = 32
	var wg sync.WaitGroup
	for range workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := service.WithState(func(state *lua.LState) error {
				value := state.GetGlobal("counter")
				counter := 0
				if value != lua.LNil {
					counter = int(value.(lua.LNumber))
				}
				state.SetGlobal("counter", lua.LNumber(counter+1))
				return nil
			}); err != nil {
				t.Errorf("updating state: %v", err)
			}
		}()
	}
	wg.Wait()

	if err := service.WithState(func(state *lua.LState) error {
		if got := int(state.GetGlobal("counter").(lua.LNumber)); got != workers {
			t.Fatalf("counter = %d, want %d", got, workers)
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
}
