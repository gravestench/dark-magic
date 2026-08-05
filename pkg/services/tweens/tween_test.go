package tweens

import (
	"testing"
	"time"
)

func TestTweenCompletesAtExactDuration(t *testing.T) {
	service := &Service{}
	completed := false
	lastProgress := 0.0
	tween := service.New().Time(time.Second).
		OnUpdate(func(progress float64) { lastProgress = progress }).
		OnComplete(func() { completed = true })
	tween.Update(time.Second)
	if !completed || lastProgress != 1 {
		t.Fatalf("completed=%v progress=%v", completed, lastProgress)
	}
}

func TestTweenRepeatForever(t *testing.T) {
	service := &Service{}
	repeats := 0
	tween := service.New().Time(time.Second).Repeat(RepeatForever).OnRepeat(func() { repeats++ })
	tween.Update(time.Second)
	tween.Update(time.Second)
	if repeats != 2 {
		t.Fatalf("repeats=%d, want 2", repeats)
	}
}
