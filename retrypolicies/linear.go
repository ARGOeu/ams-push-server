package retrypolicies

import "time"

// Linear implements the RetryPolicy interface
// A linear retry policy resets its timer on a fixed interval ignoring any possible error
type Linear struct {
	period time.Duration
	timer  *time.Timer
}

// Reset resets the timer based on the registered period
func (l *Linear) Reset(err string) {
	l.timer.Reset(l.period)
}

// Timer returns the in use timer of the policy
func (l *Linear) Timer() *time.Timer {
	return l.timer
}
