package utils

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
)

func MakeWhere(
	ctx context.Context,
	tbl *gorm.DB,
	params url.Values,
	useAliasForModel bool,
) *gorm.DB {
	if val, ok := params["filterModel"]; ok && len(val) > 0 && val[0] != "" {
		tableName := "payment_documents_combo"
		if useAliasForModel {
			tableName = "pd"
		}

		tbl = CreateFilter(ctx, tbl, val[0], tableName)
	}

	startStr, startParamFound := params["start"]
	endStr, endParamFound := params["end"]

	if startParamFound && endParamFound {
		start, err := time.Parse("02.01.2006 15:04:05", startStr[0])
		if err != nil {
			return tbl
		}

		end, err := time.Parse("02.01.2006 15:04:05", endStr[0])
		if err != nil {
			return tbl
		}

		tbl = tbl.Where(
			"date between ? and ?",
			start.Format("2006-01-02"),
			end.Format("2006-01-02"),
		)
	}

	if val, ok := params["accounts"]; ok && len(val) > 0 && val[0] != "" {
		if len(val) > 1 {
			tbl = tbl.Where("bank_detail_id in (?)", val)
		} else {
			tbl = tbl.Where("bank_detail_id in (?)", strings.Split(val[0], ","))
		}
	}

	if val, ok := params["direction"]; ok && len(val) == 1 && val[0] != "" {
		if val[0] == "in" {
			tbl = tbl.Having("credit = 0")
		}

		if val[0] == "out" {
			tbl = tbl.Having("debet = 0")
		}
	}

	if val, ok := params["sort"]; ok && len(val) > 0 && val[0] != "" {
		for _, v := range val {
			sort := strings.Split(v, ":")
			tbl = tbl.Order(fmt.Sprintf("%s %s", sort[0], sort[1]))
		}
	}

	if query, ok := params["query"]; ok {
		likePhrase := fmt.Sprintf("%%%s%%", query[0])

		sum, err := strconv.ParseFloat(query[0], 32)
		if err == nil {
			tbl = tbl.
				Where("summ = ?", sum, sum)
		} else {
			tbl = tbl.
				Where("(payment_purpose LIKE ?)", likePhrase, likePhrase)
		}
	}

	return tbl
}
