package rendercore

import "time"

// AnimationPlayer is a deterministic, backend-neutral animation clock. Time
// advances only when explicitly submitted by the owner.
type AnimationPlayer struct {
	durations []time.Duration
	mode      string
	frame     int
	direction int
	elapsed   time.Duration
	paused    bool
}

func NewAnimationPlayer(durations []time.Duration, mode string) *AnimationPlayer {
	if mode == "" {
		mode = "loop"
	}
	return &AnimationPlayer{durations: append([]time.Duration(nil), durations...), mode: mode, direction: 1}
}

func (p *AnimationPlayer) Frame() int { return p.frame }

func (p *AnimationPlayer) SetPaused(paused bool) { p.paused = paused }

func (p *AnimationPlayer) Seek(position time.Duration) (int, bool) {
	previous := p.frame
	p.frame, p.direction, p.elapsed = 0, 1, 0
	if position < 0 {
		position = 0
	}
	if cycle := p.cycleDuration(); cycle > 0 {
		switch p.mode {
		case "once":
			if position >= cycle {
				position = cycle - time.Nanosecond
			}
		default:
			position %= cycle
		}
	}
	p.advance(position)
	return p.frame, p.frame != previous
}

func (p *AnimationPlayer) Advance(delta time.Duration) (int, bool) {
	if p.paused || delta <= 0 || len(p.durations) == 0 {
		return p.frame, false
	}
	return p.advance(delta)
}

func (p *AnimationPlayer) advance(delta time.Duration) (int, bool) {
	previous := p.frame
	p.elapsed += delta
	for p.elapsed >= p.durations[p.frame] {
		p.elapsed -= p.durations[p.frame]
		switch p.mode {
		case "once":
			if p.frame == len(p.durations)-1 {
				p.elapsed = 0
				return p.frame, p.frame != previous
			}
			p.frame++
		case "ping-pong":
			if len(p.durations) == 1 {
				p.elapsed = 0
				return p.frame, p.frame != previous
			}
			p.frame += p.direction
			if p.frame == len(p.durations)-1 || p.frame == 0 {
				p.direction = -p.direction
			}
		default:
			p.frame = (p.frame + 1) % len(p.durations)
		}
	}
	return p.frame, p.frame != previous
}

func (p *AnimationPlayer) cycleDuration() time.Duration {
	var total time.Duration
	for _, duration := range p.durations {
		total += duration
	}
	if p.mode == "ping-pong" && len(p.durations) > 1 {
		return 2*total - p.durations[0] - p.durations[len(p.durations)-1]
	}
	return total
}
