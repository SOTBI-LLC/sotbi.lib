package onec

import (
	"time"

	pb "github.com/SOTBI-LLC/sotbi.lib/pkg/api/onec"
)

type AccountBalance struct {
	ID             uint64     `json:"id"`
	ExchangeFileID uint64     `json:"exchange_file_id"`
	StartDateStr   string     `json:"-"                    mapstructure:"ДатаНачала"`
	StartDate      *time.Time `json:"start_date,omitempty" mapstructure:"-"`
	EndDateStr     string     `json:"-"                    mapstructure:"ДатаКонца"`
	EndDate        *time.Time `json:"end_date,omitempty"   mapstructure:"-"`
	Account        string     `json:"account,omitempty"    mapstructure:"РасчСчет"`
	InitialBalance float64    `json:"initial_balance"      mapstructure:"НачальныйОстаток"`
	Income         float64    `json:"income"               mapstructure:"ВсегоПоступило"`
	WriteOff       float64    `json:"write_off"            mapstructure:"ВсегоСписано"`
	FinalBalance   float64    `json:"final_balance"        mapstructure:"КонечныйОстаток"`
}

func (b *AccountBalance) ToPB(request *pb.ParseRequest) *pb.ParseResponse {
	return &pb.ParseResponse{
		RequestId:    request.RequestId,
		CustomerType: request.CustomerType,
		FileUrl:      request.FileUrl,
		CreatorId:    request.CreatorId,
		DebtorId:     request.DebtorId,
		Item: &pb.ParseResponse_Balance{
			Balance: &pb.AccountBalance{
				Id:             b.ID,
				StartDate:      timeToTimestamppb(b.StartDate),
				EndDate:        timeToTimestamppb(b.EndDate),
				Account:        b.Account,
				InitialBalance: b.InitialBalance,
				Income:         b.Income,
				WriteOff:       b.WriteOff,
				FinalBalance:   b.FinalBalance,
				ExchangeFileId: b.ExchangeFileID,
			},
		},
	}
}
