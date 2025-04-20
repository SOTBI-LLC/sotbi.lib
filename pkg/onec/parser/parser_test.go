package parser

import (
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/suite"
)

type ParserTestSuite struct {
	suite.Suite
}

func (suite *ParserTestSuite) TestScan() {
	file, err := os.Open("fixtures/0.txt")
	suite.NoError(err)
	defer file.Close()

	outputFile, err := os.Create("fixtures/out.txt")
	suite.NoError(err)
	defer outputFile.Close()

	p := &ExchangeFile{}

	r := io.TeeReader(file, outputFile)

	// Call the Scan function
	result, err := p.Scan(r)

	// Assert that there is no error
	suite.NoError(err)

	// Assert that the result is not nil
	suite.NotNil(result)

	io.Copy(file, outputFile)

	outputFile.Close()
}

func TestParserTestSuite(t *testing.T) {
	suite.Run(t, new(ParserTestSuite))
}
