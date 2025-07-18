package onec

import (
	"io"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"
)

// Helper functions.
func ParseDateTime(date string) *time.Time {
	if date == "" {
		return nil
	}

	loc, err := time.LoadLocation("Europe/Moscow")
	if err != nil {
		loc = time.FixedZone("UTC+3", 3*60*60)
	}

	t, err := time.ParseInLocation("02.01.2006 15:04:05", date, loc)
	if err != nil {
		t, err = time.ParseInLocation("2006-01-02 15:04:05", date, loc)
		if err != nil {
			t, err = time.ParseInLocation("02012006 15:04:05", date, loc)
			if err != nil {
				t, err = time.ParseInLocation("2006/01/02 15:04:05", date, loc)
				if err != nil {
					return nil
				}
			}
		}
	}

	return &t
}

func ParseDate(date string) *time.Time {
	if date == "" {
		return nil
	}

	t, err := time.Parse("02.01.2006", date)
	if err != nil {
		t, err = time.Parse("2006-01-02", date)
		if err != nil {
			t, err = time.Parse("02012006", date)
			if err != nil {
				t, err = time.Parse("2006/01/02", date)
				if err != nil {
					return nil
				}
			}
		}
	}

	return &t
}

func timeToTimestamppb(t *time.Time) *timestamppb.Timestamp {
	if t == nil {
		return nil
	}

	return timestamppb.New(*t)
}

type Result struct {
	ExchangeFile     ExchangeFile
	Remainings       []AccountBalance
	PaymentDocuments []PaymentDocument
}

type Parser interface {
	Scan(io.Reader, func() (uint64, error)) (*Result, error)
}

// ProcessBalanceAndDocs optimized version using maps for O(n+m) complexity.
func (r *Result) ProcessBalanceAndDocs() *Result {
	if len(r.Remainings) == 0 || len(r.PaymentDocuments) == 0 {
		return r
	}

	// Group remainings by account for fast lookup
	remainingsByAccount := make(map[string][]AccountBalance, len(r.Remainings))
	for _, remaining := range r.Remainings {
		remainingsByAccount[remaining.Account] = append(
			remainingsByAccount[remaining.Account],
			remaining,
		)
	}

	// Process each payment document once
	for j := range r.PaymentDocuments {
		doc := &r.PaymentDocuments[j]

		// Check payer account first
		if doc.WrittenOffDate != nil {
			if remainings, exists := remainingsByAccount[doc.PayerAccount]; exists {
				for _, remaining := range remainings {
					if isDateInRange(doc.WrittenOffDate, remaining.StartDate, remaining.EndDate) {
						r.PaymentDocuments[j].AccountBalanceID = remaining.ID

						goto nextDoc // Found match, skip receiver check
					}
				}
			}
		}

		// Check receiver account only if payer didn't match
		if doc.IncomeDate != nil {
			if remainings, exists := remainingsByAccount[doc.ReceiverAccount]; exists {
				for _, remaining := range remainings {
					if isDateInRange(doc.IncomeDate, remaining.StartDate, remaining.EndDate) {
						r.PaymentDocuments[j].AccountBalanceID = remaining.ID

						break
					}
				}
			}
		}

	nextDoc:
	}

	return r
}

// isDateInRange checks if date is within the range [start, end] inclusive.
func isDateInRange(date, start, end *time.Time) bool {
	if date == nil || start == nil || end == nil {
		return false
	}

	return (date.Equal(*start) || date.After(*start)) && (date.Equal(*end) || date.Before(*end))
}
