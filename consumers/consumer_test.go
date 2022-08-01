package consumers

import (
	ams "github.com/ARGOeu/ams-push-server/pkg/ams/v1"
	"github.com/stretchr/testify/suite"
	"net/http"
	"testing"
)

type ConsumerTestSuite struct {
	suite.Suite
}

// TestNew tests that the consumer factory behaves properly
func (suite *ConsumerTestSuite) TestNew() {

	// normal creation
	c, e1 := New(AmsHttpConsumerType, "/projects/p1/subscriptions/s1",
		ams.NewClient("https", "localhost", "token", 8080, new(http.Client)))
	suite.IsType(&AmsHttpConsumer{}, c)
	suite.Equal("subscription /projects/p1/subscriptions/s1 from localhost:8080", c.ResourceInfo())
	suite.Nil(e1)

	// unimplemented consumer
	_, e2 := New("unknown", "", nil)
	suite.Equal("consumer unknown not yet implemented", e2.Error())

}

func TestConsumerTestSuite(t *testing.T) {
	suite.Run(t, new(ConsumerTestSuite))
}
