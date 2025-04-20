package email

import (
	"bytes"
	"fmt"
	"os"
	"testing"

	smtpmock "github.com/mocktools/go-smtp-mock/v2"
	"github.com/stretchr/testify/suite"
)

type EmailTestSuite struct {
	suite.Suite
	sender     *Postman
	smtpServer *smtpmock.Server
}

// SetupTest runs before each test in the suite.
func (suite *EmailTestSuite) SetupTest() {
	suite.smtpServer = smtpmock.New(smtpmock.ConfigurationAttr{
		LogToStdout:       true,
		LogServerActivity: true,
	})
	err := suite.smtpServer.Start()
	suite.Require().NoError(err)

	err = os.Setenv("EMAIL_USE_TLS", "1")
	suite.Require().NoError(err)

	cfg := Config{
		User:   "mst@grp.loc",
		Server: fmt.Sprintf("localhost:%d", suite.smtpServer.PortNumber()),
	}
	suite.sender = NewSender(&cfg)
}

// TestSendEmail tests the sending of an email.
func (suite *EmailTestSuite) TestSendEmail() {
	// Use a buffer as a mock attachment
	attachment := bytes.NewBufferString("This is a test attachment")
	err := suite.sender.
		SetSubject("test").
		SetDestination([]string{"user@grp.loc"}).
		SendEmailWithAttachment(
			"Body",
			attachment,
			"test.txt",
		)
	suite.NoError(err, "Failed to send email")
}

// TestSuite runs the test suite.
func TestEmailTestSuite(t *testing.T) {
	suite.Run(t, new(EmailTestSuite))
}
