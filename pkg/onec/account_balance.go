package onec

type AccountBalance struct {
	StartDate      string  `mapstructure:"ДатаНачала"       json:"start_date,omitempty"`
	EndDate        string  `mapstructure:"ДатаКонца"        json:"end_date,omitempty"`
	Account        string  `mapstructure:"РасчСчет"         json:"account,omitempty"`
	InitialBalance float64 `mapstructure:"НачальныйОстаток" json:"initial_balance"`
	Income         float64 `mapstructure:"ВсегоПоступило"   json:"income"`
	WriteOff       float64 `mapstructure:"ВсегоСписано"     json:"write_off"`
	FinalBalance   float64 `mapstructure:"КонечныйОстаток"  json:"final_balance"`
}
