package graph

import (
	"errors"
	"time"

	"github.com/kurtschwarz/homie/services/api/internal/graph/model"
)

type TimerState struct {
	Title           string
	Emoji           string
	DurationSeconds int
	EndsAt          time.Time
	Remaining       time.Duration
	Paused          bool
}

func (r *mutationResolver) updateTimer(update func(*TimerState, time.Time)) (*model.Timer, error) {
	r.Resolver.TimerMu.Lock()
	defer r.Resolver.TimerMu.Unlock()
	if r.Resolver.Timer == nil {
		return nil, errors.New("no active timer")
	}
	now := time.Now()
	update(r.Resolver.Timer, now)
	return timerModel(r.Resolver.Timer, now), nil
}

func timerModel(timer *TimerState, now time.Time) *model.Timer {
	remaining := timer.Remaining
	if !timer.Paused {
		remaining = timer.EndsAt.Sub(now)
	}
	seconds := int(remaining.Seconds())
	if remaining > 0 && remaining%time.Second != 0 {
		seconds++
	}
	return &model.Timer{Title: timer.Title, Emoji: timer.Emoji, DurationSeconds: timer.DurationSeconds, RemainingSeconds: seconds, EndsAt: int(timer.EndsAt.Unix()), Paused: timer.Paused}
}
