package tweens

import (
	"time"

	"github.com/gravestench/dark-magic/pkg/easing"
)

const (
	RepeatForever   = -1
	defaultDuration = time.Second / 2
	defaultEase     = easing.Linear
)

type tweenConfig struct {
	duration    time.Duration
	delay       time.Duration
	justStarted bool
	repeatCount int
	easeName    string
	ease        func(complete float64) float64
	onStart     func()
	onComplete  func()
	onRepeat    func()
	onUpdate    func(complete float64)
}
