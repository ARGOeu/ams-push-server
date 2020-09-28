package retrypolicies

import (
	"github.com/stretchr/testify/suite"
	"testing"
	"time"
)

type SlowStartTestSuite struct {
	suite.Suite
}

func (suite *SlowStartTestSuite) TestReset() {

	lr := Slowstart{
		previousRestartInterval: 1 * time.Second,
		previousError:           false,
		timer:                   time.NewTimer(0),
	}

	// drain the first tick
	time.Sleep(SlowStartInitialInterval)
	t1 := <-lr.timer.C

	// reset the timer with no error
	lr.Reset("")

	// wait for the tick to happen
	time.Sleep(500 * time.Millisecond)

	// drain the second tick
	t2 := <-lr.timer.C

	// make sure that the two time events have more than the halved period (500ms) in between them
	suite.True(t2.Sub(t1) > 500*time.Millisecond)
	suite.False(lr.previousError)
	suite.Equal(500*time.Millisecond, lr.previousRestartInterval)

	// hit the lower bound
	lr.Reset("")
	time.Sleep(300 * time.Millisecond)
	t3 := <-lr.timer.C

	suite.True(t3.Sub(t2) > 300*time.Millisecond)
	suite.False(lr.previousError)
	suite.Equal(300*time.Millisecond, lr.previousRestartInterval)

	// error cases
	lr2 := Slowstart{
		previousRestartInterval: 12 * time.Hour,
		previousError:           true,
		timer:                   time.NewTimer(0),
	}

	lr2.Reset("error")
	suite.Equal(24*time.Hour, lr2.previousRestartInterval)
	suite.True(lr2.previousError)

	// don't exceed the upper bound
	lr2.Reset("error")
	suite.Equal(24*time.Hour, lr2.previousRestartInterval)
	suite.True(lr2.previousError)

	// no error should reset the timer to SlowStartInitialInterval
	lr2.Reset("")
	suite.Equal(SlowStartInitialInterval, lr2.previousRestartInterval)
	suite.False(lr2.previousError)

}

func (suite *SlowStartTestSuite) TestTimer() {

	t1 := time.NewTimer(0)

	lr := Slowstart{
		timer: t1,
	}

	suite.Equal(t1, lr.Timer())

}

func TestSlowStartTestSuite(t *testing.T) {
	suite.Run(t, new(SlowStartTestSuite))
}
