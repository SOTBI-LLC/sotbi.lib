package onec

import (
	"strings"
	"time"

	pb "github.com/SOTBI-LLC/sotbi.lib/pkg/api/onec"
)

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

func (f *ExchangeFile) ToPB(request *pb.ParseRequest) *pb.ParseResponse {
	var accounts []string
	if (len(f.Account) == 0 || f.Account[0] == "") && request.Account != nil {
		accounts = []string{*request.Account}
	} else {
		accounts = f.Account
	}

	return &pb.ParseResponse{
		RequestId:    request.RequestId,
		CustomerType: request.CustomerType,
		FileUrl:      request.FileUrl,
		CreatorId:    request.CreatorId,
		DebtorId:     request.DebtorId,
		Item: &pb.ParseResponse_File{
			File: &pb.ExchangeFile{
				Id:              f.ID,
				FormatVer:       strings.TrimSpace(f.FormatVer),
				Encoding:        strings.TrimSpace(f.Encoding),
				Sender:          strings.TrimSpace(f.Sender),
				Receiver:        strings.TrimSpace(f.Receiver),
				CreatedDatetime: timeToTimestamppb(f.CreatedDate),
				StartDate:       timeToTimestamppb(f.StartDate),
				EndDate:         timeToTimestamppb(f.EndDate),
				Account:         accounts,
			},
		},
	}
}
