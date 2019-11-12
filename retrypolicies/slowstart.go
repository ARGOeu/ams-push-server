package retrypolicies

import "time"

const (
	SlowStartInitialInterval = 1 * time.Second
	SlowStartLowerTimeBound  = 300 * time.Millisecond
	SlowStartUpperTimeBound  = 24 * time.Hour
)

// Slowstart implements the retry policy interface
// A slowstart retry policy performs faster push cycles when there are no errors and slower ones when errors exist
type Slowstart struct {
	previousRestartInterval time.Duration
	previousError           bool
	timer                   *time.Timer
}

// Reset resets the timer based on the registered period
func (s *Slowstart) Reset(err string) {

	var restartInterval time.Duration

	// if there was no error registered and there is STILL NO error,half the interval
	if !s.previousError && err == "" {
		// check to not exceed the lower bound
		if s.previousRestartInterval/2 < SlowStartLowerTimeBound {
			restartInterval = SlowStartLowerTimeBound
		} else {
			restartInterval = s.previousRestartInterval / 2
		}
	}

	// if there is an error, double the interval
	if err != "" {
		// check to not exceed the upper bound
		if s.previousRestartInterval*2 > SlowStartUpperTimeBound {
			restartInterval = SlowStartUpperTimeBound
		} else {
			restartInterval = s.previousRestartInterval * 2
		}

		s.previousError = true
	}

	// if there was previously an error that has now been resolved, reset the timer to the initial interval
	if s.previousError && err == "" {
		restartInterval = SlowStartInitialInterval
		s.previousError = false
	}

	// update the restart interval
	s.previousRestartInterval = restartInterval
	s.timer.Reset(restartInterval)

}

// Timer returns the in use timer of the policy
func (s *Slowstart) Timer() *time.Timer {
	return s.timer
}
