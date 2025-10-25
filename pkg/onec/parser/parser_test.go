package parser

import (
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/SOTBI-LLC/sotbi.lib/pkg/utils"
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

	sonyflake, err := utils.NewSonyflake(utils.SonyflakeConfig{MachineID: 1})
	suite.NoError(err)

	// Call the Scan function
	result, err := p.Scan(r, sonyflake.NextID)

	// Assert that there is no error
	suite.NoError(err)

	// Assert that the result is not nil
	suite.NotNil(result)

	// 1. Header (ExchangeFile)
	header := result.ExchangeFile
	suite.Equal("1.03", header.FormatVer)
	suite.Equal("Windows", header.Encoding)
	suite.Equal("Альфа-Бизнес Онлайн", header.Sender)
	suite.Equal("", header.Receiver)
	suite.Equal("31.12.2021", header.CreatedDateStr)
	suite.Equal("00:00:00", header.CreatedTimeStr)
	suite.Equal("01.01.2019", header.StartDateStr)
	suite.Equal("31.12.2021", header.EndDateStr)
	suite.Equal([]string{"12345678901234567890"}, header.Account)

	// 2. Account balances
	suite.Len(result.Remainings, 1)
	rab := result.Remainings[0]
	suite.Equal("01.01.2019", rab.StartDateStr)
	suite.Equal("31.12.2021", rab.EndDateStr)
	suite.Equal("12345678901234567890", rab.Account)
	suite.Equal(10.0, rab.InitialBalance)
	suite.Equal(2.0, rab.Income)
	suite.Equal(1.0, rab.WriteOff)
	suite.Equal(11.0, rab.FinalBalance)

	// 3. Payment documents
	suite.Len(result.PaymentDocuments, 2)

	// First document: Банковский ордер
	d1 := result.PaymentDocuments[0]
	suite.Equal("Банковский ордер", d1.DocumentType)
	suite.Equal("1", d1.Number)
	suite.Equal("01.02.2021", d1.DataStr)
	suite.Equal(2.0, d1.Summ)
	suite.Equal("12345678901234567890", d1.PayerAccount)
	suite.Equal(
		"ОБЩЕСТВО С ОГРАНИЧЕННОЙ ОТВЕТСТВЕННОСТЬЮ \"РОГА И КОПЫТА\" (ФИРМА РОГА И КОПЫТА LTD.)",
		d1.Payer,
	)
	suite.Equal("7706095014", d1.PayerINN)
	suite.Nil(d1.Payer1)
	suite.Equal(
		"Ком-я за внеш.переводы в валюте РФ на сч.ФЛ за 31ЯНВ21 Согласно тарифам Банка ООО ФИРМА РОГА И КОПЫТА LTD.",
		d1.PaymentPurpose,
	)

	// Second document: Платежное поручение
	d2 := result.PaymentDocuments[1]
	suite.Equal("Платежное поручение", d2.DocumentType)
	suite.Equal("2", d2.Number)
	suite.Equal("01.03.2021", d2.DataStr)
	suite.Equal(1.0, d2.Summ)
	suite.Equal("12345678901234567890", d2.PayerAccount)
	suite.Equal(
		"ОБЩЕСТВО С ОГРАНИЧЕННОЙ ОТВЕТСТВЕННОСТЬЮ \"ФИРМА РОГА И КОПЫТА\" (ФИРМА ФИРМА РОГА И КОПЫТА LTD.)",
		d2.Payer,
	)
	suite.Equal("7712345678", d2.PayerINN)
	suite.Equal("Иванов Иван Иванович", d2.Receiver)
	suite.Equal("Вознаграждение за январь 2021 года. НДС не облагается", d2.PaymentPurpose)

	io.Copy(file, outputFile)

	outputFile.Close()
}

func TestParserTestSuite(t *testing.T) {
	suite.Run(t, new(ParserTestSuite))
}
