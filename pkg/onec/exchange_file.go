package onec

import "time"

type ExchangeFile struct {
	ID             uint64     `json:"id"`
	FormatVer      string     `json:"format_ver"             mapstructure:"ВерсияФормата"`
	Encoding       string     `json:"encoding"               mapstructure:"Кодировка"`
	Sender         string     `json:"sender"                 mapstructure:"Отправитель"`
	Receiver       string     `json:"receiver"               mapstructure:"Получатель"`
	CreatedDateStr string     `json:"-"                      mapstructure:"ДатаСоздания"`
	CreatedTimeStr string     `json:"-"                      mapstructure:"ВремяСоздания"`
	CreatedDate    *time.Time `json:"created_date,omitempty" mapstructure:"-"`
	StartDateStr   string     `json:"-"                      mapstructure:"ДатаНачала"`
	StartDate      *time.Time `json:"start_date,omitempty"   mapstructure:"-"`
	EndDateStr     string     `json:"-"                      mapstructure:"ДатаКонца"`
	EndDate        *time.Time `json:"end_date,omitempty"     mapstructure:"-"`
	Account        []string   `json:"account,omitempty"      mapstructure:"РасчСчет,omitempty"`
}
