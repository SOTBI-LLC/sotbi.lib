package onec

type ExchangeFile struct {
	FormatVer   string   `mapstructure:"ВерсияФормата"      json:"format_ver"`
	Encoding    string   `mapstructure:"Кодировка"          json:"encoding"`
	Sender      string   `mapstructure:"Отправитель"        json:"sender"`
	Receiver    string   `mapstructure:"Получатель"         json:"receiver"`
	CreatedDate string   `mapstructure:"ДатаСоздания"       json:"created_date,omitempty"`
	CreatedTime string   `mapstructure:"ВремяСоздания"      json:"created_time,omitempty"`
	StartDate   string   `mapstructure:"ДатаНачала"         json:"start_date,omitempty"`
	EndDate     string   `mapstructure:"ДатаКонца"          json:"end_date,omitempty"`
	Account     []string `mapstructure:"РасчСчет,omitempty" json:"account,omitempty"`
}
