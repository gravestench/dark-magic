package tweens

import (
	"time"

	"github.com/gravestench/dark-magic/pkg/easing"
)

type mode int

const (
	playingForward mode = iota
	paused
	finished
)

type Tween struct {
	*tweenConfig
	mode
	elapsed time.Duration
}

func (t *Tween) Start() *Tween {
	t.elapsed = 0

	t.mode = playingForward

	return t
}

func (t *Tween) Stop() *Tween {
	t.mode = paused

	return t
}

func (t *Tween) Play() *Tween {
	t.mode = playingForward

	return t
}

func (t *Tween) Pause() *Tween {
	t.mode = paused

	return t
}

func (t *Tween) Progress() float64 {
	return float64((t.elapsed - t.delay).Milliseconds()) / float64(t.duration.Milliseconds())
}

func (t *Tween) Update(dt time.Duration) *Tween {
	if t.mode == paused || t.mode == finished {
		return t
	}

	if t.justStarted {
		if t.onStart != nil {
			t.onStart()
		}

		t.justStarted = false
	}

	t.elapsed += dt

	total := (t.delay + t.duration)

	if t.elapsed > total {
		t.elapsed %= total
		if t.repeatCount > 0 {
			t.repeatCount--
			t.justStarted = true
		} else {
			if t.onComplete != nil {
				t.onComplete()
			}

			t.elapsed = t.delay + t.duration
			t.Stop()
		}
	}

	if t.elapsed < t.delay {
		return t
	}

	progress := t.Progress()

	if t.elapsed < total && t.onUpdate != nil {
		t.onUpdate(t.ease(progress))
	}

	return t
}

func (t *Tween) Time(dt time.Duration) *Tween {
	t.duration = dt

	return t
}

func (t *Tween) Ease(args ...interface{}) *Tween {
	var (
		easeFn func(float64) float64
		name   string
	)

	if len(args) >= 2 {
		if params, ok := args[1].([]float64); ok {
			easeFn, name = getEaseFn(args[0], params)
		}
	} else if len(args) == 1 {
		easeFn, name = getEaseFn(args[0], nil)
	} else {
		easeFn, name = getEaseFn(defaultEase, nil)
	}

	t.ease = easeFn
	t.easeName = name

	return t
}

func (t *Tween) OnStart(fn func()) *Tween {
	t.onStart = fn

	return t
}

func (t *Tween) OnComplete(fn func()) *Tween {
	t.onComplete = fn

	return t
}

func (t *Tween) OnRepeat(fn func()) *Tween {
	t.onRepeat = fn

	return t
}

func (t *Tween) OnUpdate(fn func(float64)) *Tween {
	t.onUpdate = fn

	return t
}

func (t *Tween) Delay(dt time.Duration) *Tween {
	t.delay = dt

	return t
}

func (t *Tween) Repeat(count int) *Tween {
	t.repeatCount = count

	return t
}

func getEaseFn(ease interface{}, params []float64) (func(float64) float64, string) {
	switch e := ease.(type) {
	case string:
		provider, found := easing.EaseMap[e]
		if found {
			return provider.New(params), e
		}
	}

	return easing.EaseMap[easing.Default].New(params), easing.Default
}
