package templates

import (
	"time"
)

var (
	location   = time.FixedZone("UTC+3", 3*60*60)
	start      = time.Date(2021, 12, 1, 0, 0, 0, 0, location)
	end        = time.Date(2023, 12, 2, 23, 59, 59, 0, location)
	TestString = "Test String"
)

var accountStatementDataNew = struct {
	ID           uint
	Status       string
	BankDetailID uint
	BankDetail   *struct {
		Debtor      *struct{ Name string }
		Bank        *string
		BankAccount string
	}
	Start         *time.Time
	End           *time.Time
	RequestType   uint
	Creator       *struct{ User string }
	CreatedAt     *time.Time
	RequestReason *string
}{
	ID:           30368,
	Status:       "open",
	BankDetailID: 51,
	BankDetail: &struct {
		Debtor      *struct{ Name string }
		Bank        *string
		BankAccount string
	}{
		Debtor:      &struct{ Name string }{Name: "Test Debtor"},
		Bank:        &TestString,
		BankAccount: TestString,
	},
	Start:         &start,
	End:           &end,
	RequestType:   2,
	Creator:       &struct{ User string }{User: "user1"},
	CreatedAt:     &start,
	RequestReason: &TestString,
}

var templateAccountStatementNew = `
Поступила новая заявка на выписку с расчётного счёта: [перейти к заявке № {{.Data.ID}}](http://{{.IP}}/requests/account-statement/{{.Data.ID}})

|                 |                                                              |
|-----------------|--------------------------------------------------------------|
| Владелец счёта: | {{.Data.BankDetail.Debtor.Name}}                             |
| Расчётный счёт: | {{.Data.BankDetail.BankAccount}} ({{.Data.BankDetail.Bank}}) |

|                  |                                                                            |
|------------------|----------------------------------------------------------------------------|
| Период:          | с {{.Data.Start.Format "02.01.2006"}} по {{.Data.End.Format "02.01.2006"}} |
| Вид выписки:     | {{getRequestFiles .Data.RequestType}}                                                      |
| **Комментарий:** | **{{.Data.RequestReason}}**                                                |

|           |                                         |
|-----------|-----------------------------------------|
| Заказчик: | {{.Data.Creator.User}}                  |
| Дата:     | {{.Data.CreatedAt.Format "02.01.2006"}} |
`

var result = `<!DOCTYPE html>`
