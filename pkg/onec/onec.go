package onec

import (
	"io"
)

type Result struct {
	ExchangeFile     ExchangeFile
	Remainings       []AccountBalance
	PaymentDocuments []PaymentDocument
}

type Parser interface {
	Scan(io.Reader) (*Result, error)
}
