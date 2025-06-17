package onec

type ExchangeFile struct {
	FormatVer   string   `mapstructure:"ВерсияФормата,omitempty" json:"format_ver"`
	Encoding    string   `mapstructure:"Кодировка,omitempty"     json:"encoding"`
	Sender      string   `mapstructure:"Отправитель,omitempty"   json:"sender"`
	Receiver    string   `mapstructure:"Получатель,omitempty"    json:"receiver"`
	CreatedDate string   `mapstructure:"ДатаСоздания,omitempty"  json:"created_date,omitempty"`
	CreatedTime string   `mapstructure:"ВремяСоздания,omitempty" json:"created_time,omitempty"`
	StartDate   string   `mapstructure:"ДатаНачала,omitempty"    json:"start_date,omitempty"`
	EndDate     string   `mapstructure:"ДатаКонца,omitempty"     json:"end_date,omitempty"`
	Account     []string `mapstructure:"РасчСчет,omitempty"      json:"account,omitempty"`
}
