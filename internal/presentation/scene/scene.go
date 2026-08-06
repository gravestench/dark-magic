// Package scene contains renderer-independent interactive scene state. Keeping
// movement and persistence headless makes the gameplay loop deterministic and
// testable before it is attached to Raylib.
package scene

import (
	"encoding/json"
	"fmt"
	"io"
)

type Point struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

type Bounds struct {
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
}

type State struct {
	Seed   uint64 `json:"seed"`
	World  Bounds `json:"world"`
	Hero   Point  `json:"hero"`
	Camera Point  `json:"camera"`
}

func New(seed uint64, width, height float64) *State {
	state := &State{Seed: seed, World: Bounds{Width: width, Height: height}}
	state.Hero = Point{X: width / 2, Y: height / 2}
	state.trackHero()
	return state
}

func (s *State) MoveHero(dx, dy float64) {
	s.Hero.X = clamp(s.Hero.X+dx, 0, s.World.Width)
	s.Hero.Y = clamp(s.Hero.Y+dy, 0, s.World.Height)
	s.trackHero()
}

func (s *State) trackHero() {
	s.Camera = s.Hero
}

func (s *State) Save(writer io.Writer) error {
	if err := json.NewEncoder(writer).Encode(s); err != nil {
		return fmt.Errorf("encoding scene: %w", err)
	}
	return nil
}

func Load(reader io.Reader) (*State, error) {
	var state State
	if err := json.NewDecoder(reader).Decode(&state); err != nil {
		return nil, fmt.Errorf("decoding scene: %w", err)
	}
	if state.World.Width <= 0 || state.World.Height <= 0 {
		return nil, fmt.Errorf("scene has invalid world bounds")
	}
	state.MoveHero(0, 0)
	return &state, nil
}

func clamp(value, minimum, maximum float64) float64 {
	if value < minimum {
		return minimum
	}
	if value > maximum {
		return maximum
	}
	return value
}
