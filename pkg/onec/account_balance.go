package onec

type AccountBalance struct {
	StartDate      string  `mapstructure:"ДатаНачала,omitempty"       json:"start_date,omitempty"`
	EndDate        string  `mapstructure:"ДатаКонца,omitempty"        json:"end_date,omitempty"`
	Account        string  `mapstructure:"РасчСчет,omitempty"         json:"account,omitempty"`
	InitialBalance float64 `mapstructure:"НачальныйОстаток,omitempty" json:"initial_balance"`
	Income         float64 `mapstructure:"ВсегоПоступило,omitempty"   json:"income"`
	WriteOff       float64 `mapstructure:"ВсегоСписано,omitempty"     json:"write_off"`
	FinalBalance   float64 `mapstructure:"КонечныйОстаток,omitempty"  json:"final_balance"`
}
