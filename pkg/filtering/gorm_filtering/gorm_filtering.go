package gorm_filtering

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"gorm.io/gorm"

	"github.com/COTBU/sotbi.lib/pkg/filtering"
)

// CreateOrder func.
func CreateOrder(tbl *gorm.DB, args ...string) *gorm.DB {
	var sortModel filtering.SortModel

	if len(args) == 0 {
		return tbl
	}

	err := json.Unmarshal([]byte(args[0]), &sortModel)
	if err != nil {
		return tbl
	}

	prefix := ""

	if len(args) > 1 && args[1] != "" {
		prefix = fmt.Sprintf("%s.", args[1])
	}

	for _, sort := range sortModel {
		if strings.Contains(sort["colId"], ".") {
			tbl = tbl.Order(fmt.Sprintf("%s %s", sort["colId"], sort["sort"]))
		} else {
			tbl = tbl.Order(fmt.Sprintf("%s%s %s", prefix, sort["colId"], sort["sort"]))
		}
	}

	return tbl
}

// CreateFilter func.
func CreateFilter(ctx context.Context, tbl *gorm.DB, args ...string) *gorm.DB {
	if len(args) == 0 {
		return tbl
	}

	filterModel, err := filtering.ParseJSONToFilterModel(args[0])
	if err != nil {
		return tbl
	}

	prefix := ""

	if len(args) > 1 && args[1] != "" {
		prefix = fmt.Sprintf("%s.", args[1])
	}

	for field, filter := range filterModel {
		if filter.IsEmpty() {
			if !strings.Contains(field, ".") {
				field = fmt.Sprintf("%s%s", prefix, field)
			}

			var filterOperator string

			if filter.Operator != nil {
				filterOperator = *filter.Operator
			}

			switch *filter.FilterType {
			case "set":
				if len(filter.Values) > 0 {
					// проверка на наличие в set null значения
					nullExistInSet := false

					for _, val := range filter.Values {
						if val == nil || *val == "" {
							nullExistInSet = true

							break
						}
					}

					if nullExistInSet { // если null есть - то добавляем условие OR _ IS NULL
						tbl = tbl.Where(
							fmt.Sprintf("%s in (?) OR %s IS NULL", field, field),
							&filter.Values,
						)
					} else {
						tbl = tbl.Where(fmt.Sprintf("%s in (?)", field), filter.Values)
					}
				}
			case "date":
				if filterOperator == "" {
					tbl = constructDateWhere(tbl, field, filterOperator, filter)
				} else {
					tbl = constructDateWhere(tbl, field, filterOperator, filter.Condition1.Filter, filter.Condition2.Filter)
				}
			case "number":
				tbl = constructNumberWhere(tbl, field, filter)
			case "text":
				if filterOperator == "" {
					tbl = constructTextWhere(tbl, field, filterOperator, filter)
				} else {
					tbl = constructTextWhere(tbl, field, filterOperator, filter.Condition1.Filter, filter.Condition2.Filter)
				}
			default:
				tbl.Logger.Info(ctx, "unknown number filterType: "+*filter.FilterType)

				return tbl
			}
		}
	}

	return tbl
}

func constructNumberWhere(tbl *gorm.DB, field string, filter filtering.Filter) *gorm.DB {
	operators := map[string]string{
		"equals":             "%s = ?",
		"notEqual":           "%s <> ?",
		"greaterThan":        "%s > ?",
		"greaterThanOrEqual": "%s >= ?",
		"lessThan":           "%s < ?",
		"lessThanOrEqual":    "%s <= ?",
		"inRange":            "%s between ? and ?",
	}
	tbl = constructQuery(
		tbl,
		operators[*filter.Type],
		"",
		field,
		"",
		fmt.Sprintf("%v", *filter.Filter),
	)

	return tbl
}

func constructDateWhere(
	tbl *gorm.DB,
	field, operator string,
	filters ...filtering.Filter,
) *gorm.DB {
	var start1, start2 string
	start1 = (*filters[0].DateFrom)[0:10]

	if operator != "" {
		start2 = (*filters[1].DateFrom)[0:10]
	}

	dateOperators := map[string]string{
		"inRange":     "DATE(%s) between ? and ?",
		"equals":      "DATE(%s) = ?",
		"greaterThan": "DATE(%s) > ?",
		"lessThan":    "DATE(%s) < ?",
		"notEqual":    "DATE(%s) <> ?",
	}

	if *filters[0].Type == "inRange" {
		end1 := (*filters[0].DateTo)[0:10]
		if operator == "" {
			tbl = tbl.Where(fmt.Sprintf("DATE(%s) between ? and ?", field), start1, end1)
		} else {
			end2 := (*filters[1].DateTo)[0:10]
			tbl = tbl.Where(fmt.Sprintf("DATE(%s) between ? and ? %s DATE(%s) between ? and ?", field, operator, field),
				start1, end1, start2, end2)
		}
	} else {
		if operator == "" {
			tbl = constructQuery(
				tbl,
				dateOperators[*filters[0].Type],
				"",
				field,
				operator,
				start1)
		} else {
			tbl = constructQuery(
				tbl,
				dateOperators[*filters[0].Type],
				dateOperators[*filters[1].Type],
				field,
				operator,
				start1,
				start2)
		}
	}

	return tbl
}

func constructTextWhere(
	tbl *gorm.DB,
	field, operator string,
	filters ...filtering.Filter,
) *gorm.DB {
	operators := map[string]string{
		"equals":      "lower(%s) = ?",
		"notEqual":    "lower(%s) <> ?",
		"contains":    "lower(%s) like ?",
		"notContains": "lower(%s) not like ?",
		"startsWith":  "lower(%s) like ?",
		"endsWith":    "lower(%s) like ?",
	}
	if operator == "" {
		tbl = constructQuery(
			tbl,
			operators[*filters[0].Type],
			"",
			field,
			operator,
			likeMix(*filters[0].Type, fmt.Sprintf("%v", *filters[0].Filter)))
	} else {
		tbl = constructQuery(
			tbl,
			operators[*filters[0].Type],
			operators[*filters[1].Type],
			field,
			operator,
			likeMix(*filters[0].Type, (*filters[0].Filter).(string)), //nolint:errcheck
			likeMix(*filters[1].Type, (*filters[1].Filter).(string))) //nolint:errcheck
	}

	return tbl
}

func likeMix(typeOperator, filter string) (result string) {
	likeOperators := map[string]string{
		"contains":    "%%%s%%",
		"notContains": "%%%s%%",
		"startsWith":  "%s%%",
		"endsWith":    "%%%s",
	}

	likePhrase, ok := likeOperators[typeOperator]
	if ok {
		return fmt.Sprintf(likePhrase, strings.ToLower(filter))
	}

	return filter
}

func constructQuery(
	tbl *gorm.DB,
	query0, query1, field, operator string,
	params ...string,
) *gorm.DB {
	if operator == "" {
		tbl = tbl.Where(fmt.Sprintf(query0, field), params[0])
	} else {
		query := fmt.Sprintf(query0, field) + " " + operator + " " + fmt.Sprintf(query1, field)
		tbl = tbl.Where(query, params[0], params[1])
	}

	return tbl
}
