package squirrel_fltering

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	sq "github.com/Masterminds/squirrel"

	"github.com/SOTBI-LLC/sotbi.lib/pkg/filtering"
)

// CreateOrder func.
func CreateOrder(query sq.SelectBuilder, args ...string) sq.SelectBuilder {
	var sortModel filtering.SortModel

	if len(args) == 0 {
		return query
	}

	err := json.Unmarshal([]byte(args[0]), &sortModel)
	if err != nil {
		return query
	}

	prefix := ""

	if len(args) > 1 && args[1] != "" {
		prefix = fmt.Sprintf("%s.", args[1])
	}

	for _, sort := range sortModel {
		if strings.Contains(sort["colId"], ".") {
			query = query.OrderBy(fmt.Sprintf("%s %s", sort["colId"], sort["sort"]))
		} else {
			query = query.OrderBy(fmt.Sprintf("%s%s %s", prefix, sort["colId"], sort["sort"]))
		}
	}

	return query
}

// CreateFilter func.
func CreateFilter(query sq.SelectBuilder, args ...string) sq.SelectBuilder {
	if len(args) == 0 {
		return query
	}

	filterModel, err := filtering.ParseJSONToFilterModel(args[0])
	if err != nil {
		return query
	}

	prefix := ""

	if len(args) > 1 && args[1] != "" {
		prefix = fmt.Sprintf("%s.", args[1])
	}

	for field, filter := range filterModel {
		if filter.IsEmpty() {
			continue
		}

		if prefix != "" && !strings.Contains(field, ".") {
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
					if *val == "" {
						nullExistInSet = true

						break
					}
				}

				if nullExistInSet { // если null есть - то добавляем условие OR _ IS NULL
					query = query.Where(
						sq.Or{
							sq.Eq{field: filter.Values},
							sq.Expr(fmt.Sprintf("%s IS NULL", field)),
						},
					)
				} else {
					query = query.Where(sq.Eq{field: filter.Values})
				}
			}
		case "date":
			if filterOperator == "" {
				query = constructDateWhere(query, field, filterOperator, filter)
			} else {
				query = constructDateWhere(query, field, filterOperator, filter.Condition1.Filter, filter.Condition2.Filter)
			}
		case "number":
			query = constructNumberWhere(query, field, filter)
		case "text":
			if filterOperator == "" {
				query = constructTextWhere(query, field, filterOperator, filter)
			} else {
				query = constructTextWhere(query, field, filterOperator, filter.Condition1.Filter, filter.Condition2.Filter)
			}
		default:
			slog.Default().Info("unknown number filterType: " + *filter.FilterType)

			return query
		}
	}

	return query
}

func constructNumberWhere(
	query sq.SelectBuilder,
	field string,
	filter filtering.Filter,
) sq.SelectBuilder {
	operators := map[string]string{
		"equals":             "%s = ?",
		"notEqual":           "%s <> ?",
		"greaterThan":        "%s > ?",
		"greaterThanOrEqual": "%s >= ?",
		"lessThan":           "%s < ?",
		"lessThanOrEqual":    "%s <= ?",
		"inRange":            "%s between ? AND ?",
	}
	query = constructQuery(
		query,
		operators[*filter.Type],
		"",
		field,
		"",
		fmt.Sprintf("%v", *filter.Filter),
	)

	return query
}

func constructDateWhere(
	query sq.SelectBuilder,
	field, operator string,
	filters ...filtering.Filter,
) sq.SelectBuilder {
	var start1, start2 string
	start1 = (*filters[0].DateFrom)[0:10]

	if operator != "" {
		start2 = (*filters[1].DateFrom)[0:10]
	}

	dateOperators := map[string]string{
		"inRange":     "DATE(%s) between ? AND ?",
		"equals":      "DATE(%s) = ?",
		"greaterThan": "DATE(%s) > ?",
		"lessThan":    "DATE(%s) < ?",
		"notEqual":    "DATE(%s) <> ?",
	}

	if *filters[0].Type == "inRange" {
		end1 := (*filters[0].DateTo)[0:10]
		if operator == "" {
			query = query.Where(fmt.Sprintf("DATE(%s) between ? AND ?", field), start1, end1)
		} else {
			end2 := (*filters[1].DateTo)[0:10]
			query = query.Where(fmt.Sprintf("DATE(%s) between ? AND ? %s DATE(%s) between ? AND ?", field, operator, field),
				start1, end1, start2, end2)
		}
	} else {
		if operator == "" {
			query = constructQuery(
				query,
				dateOperators[*filters[0].Type],
				"",
				field,
				operator,
				start1)
		} else {
			query = constructQuery(
				query,
				dateOperators[*filters[0].Type],
				dateOperators[*filters[1].Type],
				field,
				operator,
				start1,
				start2)
		}
	}

	return query
}

func constructTextWhere(
	query sq.SelectBuilder,
	field, operator string,
	filters ...filtering.Filter,
) sq.SelectBuilder {
	operators := map[string]string{
		"equals":      "lower(%s) = ?",
		"notEqual":    "lower(%s) <> ?",
		"contains":    "lower(%s) LIKE ?",
		"notContains": "lower(%s) NOT LIKE ?",
		"startsWith":  "lower(%s) LIKE ?",
		"endsWith":    "lower(%s) LIKE ?",
	}
	if operator == "" {
		query = constructQuery(
			query,
			operators[*filters[0].Type],
			"",
			field,
			operator,
			likeMix(*filters[0].Type, fmt.Sprintf("%v", *filters[0].Filter)))
	} else {
		query = constructQuery(
			query,
			operators[*filters[0].Type],
			operators[*filters[1].Type],
			field,
			operator,
			likeMix(*filters[0].Type, (*filters[0].Filter).(string)), //nolint:errcheck
			likeMix(*filters[1].Type, (*filters[1].Filter).(string))) //nolint:errcheck
	}

	return query
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
	query sq.SelectBuilder,
	query0, query1, field, operator string,
	params ...string,
) sq.SelectBuilder {
	if operator == "" {
		query = query.Where(fmt.Sprintf(query0, field), params[0])
	} else {
		concatenatedQuery := fmt.Sprintf(query0, field) + " " + operator + " " + fmt.Sprintf(query1, field)
		query = query.Where(concatenatedQuery, params[0], params[1])
	}

	return query
}
