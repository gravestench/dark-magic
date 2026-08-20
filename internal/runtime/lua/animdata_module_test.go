package modruntime

import (
	"context"
	"encoding/binary"
	"testing"
	"testing/fstest"

	lua "github.com/yuin/gopher-lua"
)

type animDataFixtureReader []byte

// Read parses read at the package boundary so malformed input fails before state publication.
func (reader animDataFixtureReader) Read(string) ([]byte, error) {
	return append([]byte(nil), reader...), nil
}

// fixtureAnimData owns the fixture anim data step at this boundary, keeping its side effects and failure point
// explicit to callers.
func fixtureAnimData(name string, frames, speed int, events map[int]byte) []byte {
	const blocks, eventBytes = 256, 144

	data := make([]byte, 0, blocks*4+160)
	for block := 0; block < blocks; block++ {
		count := uint32(0)
		if block == 0 {
			count = 1
		}

		word := make([]byte, 4)
		binary.LittleEndian.PutUint32(word, count)
		data = append(data, word...)

		if count == 0 {
			continue
		}

		encodedName := make([]byte, 8)
		copy(encodedName, name)
		data = append(data, encodedName...)

		binary.LittleEndian.PutUint32(word, uint32(frames))
		data = append(data, word...)
		half := make([]byte, 2)
		binary.LittleEndian.PutUint16(half, uint16(speed))
		data = append(data, half...)
		data = append(data, 0, 0)

		encodedEvents := make([]byte, eventBytes)
		for frame, event := range events {
			encodedEvents[frame] = event
		}

		data = append(data, encodedEvents...)
	}

	return data
}

// TestAnimDataModuleExposesTypedDeterministicTimeline protects the anim data module exposes typed deterministic
// timeline contract, including its observable ordering and failure behavior.
func TestAnimDataModuleExposesTypedDeterministicTimeline(t *testing.T) {
	t.Parallel()

	source := fstest.MapFS{
		"test.lua": {
			Data: []byte(
				`local a=require("engine.animdata/v1"); timing=assert(a.record("ama1hth"))`,
			),
		},
	}

	runtime := New()
	if err := runtime.RegisterModule(
		AnimDataModule(
			animDataFixtureReader(fixtureAnimData("AMA1HTH", 8, 128, map[int]byte{3: 1, 5: 3})),
		),
	); err != nil {
		t.Fatal(err)
	}

	if err := runtime.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = runtime.Stop(context.Background()) }()

	if err := runtime.Execute(context.Background(), source, "test.lua"); err != nil {
		t.Fatal(err)
	}

	if err := runtime.Run(context.Background(), func(state *lua.LState) error {
		timing := state.GetGlobal("timing").(*lua.LTable)

		events := timing.RawGetString("events").(*lua.LTable)
		if timing.RawGetString("name").String() != "AMA1HTH" ||
			timing.RawGetString("complete_delay") != lua.LNumber(16) ||
			events.RawGetInt(1).(*lua.LTable).RawGetString("kind").String() != "attack" ||
			events.RawGetInt(1).(*lua.LTable).RawGetString("delay") != lua.LNumber(6) {
			t.Fatalf("timing = %s", timing)
		}

		return nil
	}); err != nil {
		t.Fatal(err)
	}
}
