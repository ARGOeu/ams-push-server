package retrypolicies

import (
	"github.com/stretchr/testify/suite"
	"testing"
	"time"
)

type LinearTestSuite struct {
	suite.Suite
}

func (suite *LinearTestSuite) TestReset() {

	lr := Linear{
		period: time.Duration(1000 * time.Millisecond),
		timer:  time.NewTimer(0),
	}

	// drain the first tick
	t1 := <-lr.timer.C

	// reset the timer
	lr.Reset("")

	// wait for the tick to happen
	time.Sleep(1000 * time.Millisecond)

	// drain the second tick
	t2 := <-lr.timer.C

	// make sure that the two time events have more than the configured period (1000ms) in between them
	suite.True(t2.Sub(t1) > 1000*time.Millisecond)
}

func (suite *LinearTestSuite) TestTimer() {

	t1 := time.NewTimer(0)

	lr := Linear{
		period: time.Duration(1000 * time.Millisecond),
		timer:  t1,
	}

	suite.Equal(t1, lr.Timer())

}

func TestLinearTestSuite(t *testing.T) {
	suite.Run(t, new(LinearTestSuite))
}
