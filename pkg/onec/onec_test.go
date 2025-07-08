package onec

import (
	"testing"
	"time"
)

func TestProcessBalanceAndDocs(t *testing.T) {
	// Helper function to create a time pointer
	timePtr := func(year, month, day int) *time.Time {
		t := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
		return &t
	}

	tests := []struct {
		name             string
		remainings       []AccountBalance
		paymentDocuments []PaymentDocument
		expectedMatches  map[int]uint64 // document index -> expected AccountBalanceID
	}{
		{
			name: "basic payer account matching",
			remainings: []AccountBalance{
				{
					ID:        1,
					Account:   "40702810001234567890",
					StartDate: timePtr(2024, 1, 1),
					EndDate:   timePtr(2024, 1, 31),
				},
				{
					ID:        2,
					Account:   "40702810009876543210",
					StartDate: timePtr(2024, 1, 1),
					EndDate:   timePtr(2024, 1, 31),
				},
			},
			paymentDocuments: []PaymentDocument{
				{PayerAccount: "40702810001234567890", WrittenOffDate: timePtr(2024, 1, 15)},
				{PayerAccount: "40702810009876543210", WrittenOffDate: timePtr(2024, 1, 20)},
			},
			expectedMatches: map[int]uint64{0: 1, 1: 2},
		},
		{
			name: "basic receiver account matching",
			remainings: []AccountBalance{
				{
					ID:        3,
					Account:   "40702810001234567890",
					StartDate: timePtr(2024, 1, 1),
					EndDate:   timePtr(2024, 1, 31),
				},
			},
			paymentDocuments: []PaymentDocument{
				{ReceiverAccount: "40702810001234567890", IncomeDate: timePtr(2024, 1, 15)},
			},
			expectedMatches: map[int]uint64{0: 3},
		},
		{
			name: "payer takes precedence over receiver",
			remainings: []AccountBalance{
				{
					ID:        4,
					Account:   "40702810001234567890",
					StartDate: timePtr(2024, 1, 1),
					EndDate:   timePtr(2024, 1, 31),
				},
			},
			paymentDocuments: []PaymentDocument{
				{
					PayerAccount:    "40702810001234567890",
					ReceiverAccount: "40702810001234567890",
					WrittenOffDate:  timePtr(2024, 1, 15),
					IncomeDate:      timePtr(2024, 1, 15),
				},
			},
			expectedMatches: map[int]uint64{0: 4},
		},
		{
			name: "date range boundary tests",
			remainings: []AccountBalance{
				{
					ID:        5,
					Account:   "40702810001234567890",
					StartDate: timePtr(2024, 1, 1),
					EndDate:   timePtr(2024, 1, 31),
				},
			},
			paymentDocuments: []PaymentDocument{
				{
					PayerAccount:   "40702810001234567890",
					WrittenOffDate: timePtr(2024, 1, 1),
				}, // start date
				{
					PayerAccount:   "40702810001234567890",
					WrittenOffDate: timePtr(2024, 1, 31),
				}, // end date
				{
					PayerAccount:   "40702810001234567890",
					WrittenOffDate: timePtr(2023, 12, 31),
				}, // before range
				{
					PayerAccount:   "40702810001234567890",
					WrittenOffDate: timePtr(2024, 2, 1),
				}, // after range
			},
			expectedMatches: map[int]uint64{0: 5, 1: 5}, // only first two should match
		},
		{
			name: "multiple remainings for same account",
			remainings: []AccountBalance{
				{
					ID:        6,
					Account:   "40702810001234567890",
					StartDate: timePtr(2024, 1, 1),
					EndDate:   timePtr(2024, 1, 31),
				},
				{
					ID:        7,
					Account:   "40702810001234567890",
					StartDate: timePtr(2024, 2, 1),
					EndDate:   timePtr(2024, 2, 29),
				},
			},
			paymentDocuments: []PaymentDocument{
				{PayerAccount: "40702810001234567890", WrittenOffDate: timePtr(2024, 1, 15)},
				{PayerAccount: "40702810001234567890", WrittenOffDate: timePtr(2024, 2, 15)},
			},
			expectedMatches: map[int]uint64{0: 6, 1: 7},
		},
		{
			name: "no matching account",
			remainings: []AccountBalance{
				{
					ID:        8,
					Account:   "40702810001234567890",
					StartDate: timePtr(2024, 1, 1),
					EndDate:   timePtr(2024, 1, 31),
				},
			},
			paymentDocuments: []PaymentDocument{
				{PayerAccount: "40702810009999999999", WrittenOffDate: timePtr(2024, 1, 15)},
			},
			expectedMatches: map[int]uint64{}, // no matches
		},
		{
			name: "nil dates handling",
			remainings: []AccountBalance{
				{
					ID:        9,
					Account:   "40702810001234567890",
					StartDate: timePtr(2024, 1, 1),
					EndDate:   timePtr(2024, 1, 31),
				},
			},
			paymentDocuments: []PaymentDocument{
				{PayerAccount: "40702810001234567890", WrittenOffDate: nil},
				{ReceiverAccount: "40702810001234567890", IncomeDate: nil},
			},
			expectedMatches: map[int]uint64{}, // no matches due to nil dates
		},
		{
			name:       "empty collections",
			remainings: []AccountBalance{},
			paymentDocuments: []PaymentDocument{
				{PayerAccount: "40702810001234567890", WrittenOffDate: timePtr(2024, 1, 15)},
			},
			expectedMatches: map[int]uint64{}, // no matches
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := &Result{
				Remainings:       tt.remainings,
				PaymentDocuments: tt.paymentDocuments,
			}

			result.ProcessBalanceAndDocs()

			// Check expected matches
			for docIndex, expectedID := range tt.expectedMatches {
				if docIndex >= len(result.PaymentDocuments) {
					t.Errorf("Document index %d out of range", docIndex)
					continue
				}

				actualID := result.PaymentDocuments[docIndex].AccountBalanceID
				if actualID != expectedID {
					t.Errorf(
						"Document %d: expected AccountBalanceID %d, got %d",
						docIndex,
						expectedID,
						actualID,
					)
				}
			}

			// Check that non-matching documents have AccountBalanceID = 0
			for i, doc := range result.PaymentDocuments {
				if expectedID, exists := tt.expectedMatches[i]; !exists {
					if doc.AccountBalanceID != 0 {
						t.Errorf(
							"Document %d: expected no match (AccountBalanceID=0), got %d",
							i,
							doc.AccountBalanceID,
						)
					}
				} else if doc.AccountBalanceID != expectedID {
					t.Errorf("Document %d: expected AccountBalanceID %d, got %d", i, expectedID, doc.AccountBalanceID)
				}
			}
		})
	}
}

func TestIsDateInRange(t *testing.T) {
	timePtr := func(year, month, day int) *time.Time {
		t := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
		return &t
	}

	tests := []struct {
		name     string
		date     *time.Time
		start    *time.Time
		end      *time.Time
		expected bool
	}{
		{"date in range", timePtr(2024, 1, 15), timePtr(2024, 1, 1), timePtr(2024, 1, 31), true},
		{"date equals start", timePtr(2024, 1, 1), timePtr(2024, 1, 1), timePtr(2024, 1, 31), true},
		{"date equals end", timePtr(2024, 1, 31), timePtr(2024, 1, 1), timePtr(2024, 1, 31), true},
		{
			"date before range",
			timePtr(2023, 12, 31),
			timePtr(2024, 1, 1),
			timePtr(2024, 1, 31),
			false,
		},
		{"date after range", timePtr(2024, 2, 1), timePtr(2024, 1, 1), timePtr(2024, 1, 31), false},
		{"nil date", nil, timePtr(2024, 1, 1), timePtr(2024, 1, 31), false},
		{"nil start", timePtr(2024, 1, 15), nil, timePtr(2024, 1, 31), false},
		{"nil end", timePtr(2024, 1, 15), timePtr(2024, 1, 1), nil, false},
		{"all nil", nil, nil, nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isDateInRange(tt.date, tt.start, tt.end)
			if result != tt.expected {
				t.Errorf("isDateInRange() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func BenchmarkProcessBalanceAndDocs(b *testing.B) {
	timePtr := func(year, month, day int) *time.Time {
		t := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
		return &t
	}

	// Create test data
	remainings := make([]AccountBalance, 100)
	for i := 0; i < 100; i++ {
		remainings[i] = AccountBalance{
			ID:        uint64(i + 1),
			Account:   "4070281000123456789" + string(rune('0'+i%10)),
			StartDate: timePtr(2024, 1, 1),
			EndDate:   timePtr(2024, 1, 31),
		}
	}

	paymentDocuments := make([]PaymentDocument, 1000)
	for i := 0; i < 1000; i++ {
		paymentDocuments[i] = PaymentDocument{
			PayerAccount:   "4070281000123456789" + string(rune('0'+i%10)),
			WrittenOffDate: timePtr(2024, 1, i%28+1),
		}
	}

	result := &Result{
		Remainings:       remainings,
		PaymentDocuments: paymentDocuments,
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Reset AccountBalanceID for each iteration
		for j := range result.PaymentDocuments {
			result.PaymentDocuments[j].AccountBalanceID = 0
		}

		result.ProcessBalanceAndDocs()
	}
}

func BenchmarkProcessBalanceAndDocsLarge(b *testing.B) {
	timePtr := func(year, month, day int) *time.Time {
		t := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
		return &t
	}

	// Create larger test data
	remainings := make([]AccountBalance, 1000)
	for i := 0; i < 1000; i++ {
		remainings[i] = AccountBalance{
			ID:        uint64(i + 1),
			Account:   "4070281000123456789" + string(rune('0'+i%10)),
			StartDate: timePtr(2024, 1, 1),
			EndDate:   timePtr(2024, 1, 31),
		}
	}

	paymentDocuments := make([]PaymentDocument, 10000)
	for i := 0; i < 10000; i++ {
		paymentDocuments[i] = PaymentDocument{
			PayerAccount:   "4070281000123456789" + string(rune('0'+i%10)),
			WrittenOffDate: timePtr(2024, 1, i%28+1),
		}
	}

	result := &Result{
		Remainings:       remainings,
		PaymentDocuments: paymentDocuments,
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Reset AccountBalanceID for each iteration
		for j := range result.PaymentDocuments {
			result.PaymentDocuments[j].AccountBalanceID = 0
		}

		result.ProcessBalanceAndDocs()
	}
}
