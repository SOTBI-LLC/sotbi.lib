package parser

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/mitchellh/mapstructure"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"

	"github.com/COTBU/sotbi.lib/pkg/onec"
	"github.com/COTBU/sotbi.lib/pkg/utils"
)

type ExchangeFile struct {
	exchangeFile     map[string]string
	accountBalance   []map[string]string
	paymentDocuments []map[string]string
}

var _ onec.Parser = (*ExchangeFile)(nil)

func (p *ExchangeFile) Scan(file io.Reader, next func() (uint64, error)) (*onec.Result, error) {
	if err := p.read(p.convertFileEncoding(file)); err != nil {
		return nil, err
	}

	exFile, err := p.convertFile()
	if err != nil {
		return nil, err
	}

	exFile.ID, err = next()
	if err != nil {
		return nil, err
	}

	rem, err := p.convertAccountBalance()
	if err != nil {
		return nil, err
	}

	for i := range rem {
		rem[i].ExchangeFileID = exFile.ID

		rem[i].ID, err = next()
		if err != nil {
			return nil, err
		}
	}

	pd, err := p.convertPaymentDocuments()
	if err != nil {
		return nil, err
	}

	result := &onec.Result{
		ExchangeFile:     exFile,
		Remainings:       rem,
		PaymentDocuments: pd,
	}

	return result.ProcessBalanceAndDocs(), nil
}

func (*ExchangeFile) convertFileEncoding(file io.Reader) io.Reader {
	return transform.NewReader(file, charmap.Windows1251.NewDecoder())
}

func (p *ExchangeFile) read(file io.Reader) error {
	scanner := bufio.NewScanner(file)

	const maxCapacity = 1024 * 1024 * 40 // 40MB эмпирический размер на файл
	buf := make([]byte, 0, maxCapacity)
	scanner.Buffer(buf, maxCapacity)

	var (
		currentSection map[string]string
		inSectionType  int
	)

	for scanner.Scan() {
		line := scanner.Text()

		switch {
		case strings.HasPrefix(line, "1CClientBankExchange"):
			p.exchangeFile = make(map[string]string)
			currentSection = p.exchangeFile
			inSectionType = 1

		case strings.HasPrefix(line, "СекцияРасчСчет"):
			currentSection = make(map[string]string)
			p.accountBalance = append(p.accountBalance, currentSection)
			inSectionType = 2

		case strings.HasPrefix(line, "СекцияДокумент"):
			parts := strings.SplitN(line, "=", 2)
			docType := ""

			if len(parts) == 2 {
				docType = parts[1]
			}

			currentSection = make(map[string]string)
			currentSection["СекцияДокумент"] = docType
			p.paymentDocuments = append(p.paymentDocuments, currentSection)
			inSectionType = 3

		case strings.HasPrefix(line, "КонецРасчСчет"), strings.HasPrefix(line, "КонецДокумента"):
			inSectionType = 0

		case strings.HasPrefix(line, "КонецФайла"):
			return nil

		default:
			if inSectionType != 0 && currentSection != nil {
				keyVal := strings.SplitN(strings.TrimSpace(line), "=", 2)
				if val, exist := currentSection[keyVal[0]]; len(keyVal) == 2 &&
					inSectionType == 1 &&
					exist {
					currentSection[keyVal[0]] = val + "," + keyVal[1]

					continue
				}

				if len(keyVal) == 2 {
					currentSection[keyVal[0]] = keyVal[1]
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		if errors.Is(err, bufio.ErrTooLong) {
			return fmt.Errorf("file is too long")
		}

		return fmt.Errorf("error reading data: %w", err)
	}

	return nil
}

func (p *ExchangeFile) convertFile() (onec.ExchangeFile, error) {
	var exFile onec.ExchangeFile
	config := &mapstructure.DecoderConfig{
		WeaklyTypedInput: true,
		Result:           &exFile,
	}

	decoder, err := mapstructure.NewDecoder(config)
	if err != nil {
		return exFile, fmt.Errorf("error creating new mapstructure decoder: %w", err)
	}

	if err := decoder.Decode(p.exchangeFile); err != nil {
		return exFile, fmt.Errorf("error while decoding ExchangeFile: %w", err)
	}

	exFile.StartDate = onec.ParseDate(exFile.StartDateStr)
	exFile.EndDate = onec.ParseDateTime(exFile.EndDateStr + " 23:59:59")
	exFile.CreatedDate = onec.ParseDateTime(exFile.CreatedDateStr + " " + exFile.CreatedTimeStr)

	return exFile, nil
}

func (p *ExchangeFile) convertAccountBalance() ([]onec.AccountBalance, error) {
	remainings := make([]onec.AccountBalance, 0, len(p.accountBalance))

	for i := range p.accountBalance {
		var remaining onec.AccountBalance
		config := &mapstructure.DecoderConfig{
			WeaklyTypedInput: true,
			Result:           &remaining,
		}

		decoder, err := mapstructure.NewDecoder(config)
		if err != nil {
			return nil, fmt.Errorf("error creating new mapstructure decode: %w", err)
		}

		if err := decoder.Decode(p.accountBalance[i]); err != nil {
			return nil, fmt.Errorf("error while decode remaining: %w", err)
		}

		remaining.StartDate = onec.ParseDate(remaining.StartDateStr)
		remaining.EndDate = onec.ParseDateTime(remaining.EndDateStr + " 23:59:59")

		remainings = append(remainings, remaining)
	}

	return remainings, nil
}

func (p *ExchangeFile) convertPaymentDocuments() ([]onec.PaymentDocument, error) {
	paymentDocuments := make([]onec.PaymentDocument, 0, len(p.paymentDocuments))
	for i := range p.paymentDocuments {
		var pd onec.PaymentDocument
		config := &mapstructure.DecoderConfig{
			WeaklyTypedInput: true,
			Result:           &pd,
		}

		decoder, err := mapstructure.NewDecoder(config)
		if err != nil {
			return nil, fmt.Errorf("error creating new mapstructure decode: %w", err)
		}

		if err := decoder.Decode(p.paymentDocuments[i]); err != nil {
			return nil, fmt.Errorf("error while decode payment document: %w", err)
		}

		if len(pd.RectDateStr) == 5 {
			pd.RectDateStr += ":00"
		}

		pd.Data = onec.ParseDate(pd.DataStr)
		pd.WrittenOffDate = onec.ParseDate(pd.WrittenOffDateStr)
		pd.IncomeDate = onec.ParseDate(pd.IncomeDateStr)
		pd.RectDateTime = onec.ParseDateTime(pd.RectDateStr + " " + pd.RectTimeStr)
		pd.IndicatorDate = onec.ParseDate(pd.IndicatorDateStr)
		pd.DocumentSendingDate = onec.ParseDateTime(utils.FromPtr(pd.DocumentSendingDateStr))

		paymentDocuments = append(paymentDocuments, pd)
	}

	return paymentDocuments, nil
}
