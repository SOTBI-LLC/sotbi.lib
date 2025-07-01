package raw_filtering

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/COTBU/sotbi.lib/pkg/filtering"
)

func CreateFilter(fm filtering.FilterModel, prefix *string) []string {
	out := make([]string, 0, len(fm))

	for field, f := range fm {
		if f.IsEmpty() {
			continue
		}

		if prefix != nil && *prefix != "" {
			field = *prefix + "." + field
		}

		switch t := safeDeref(f.FilterType); t {
		case "set":
			if where := setWhere(field, f.Values); where != "" {
				out = append(out, where)
			}
		case "date":
			out = append(out, dateWhere(field, f))
		case "number":
			out = append(out, numberWhere(field, f))
		case "text":
			out = append(out, textWhere(field, f))
		default:
			slog.Error("unknown filterType", "filterType", t)
		}
	}

	return out
}

// ── helpers & maps ────────────────────────────────────────────────────

var numberOps = map[string]string{
	"equals":             "%s = %v",
	"notEqual":           "%s <> %v",
	"greaterThan":        "%s > %v",
	"greaterThanOrEqual": "%s >= %v",
	"lessThan":           "%s < %v",
	"lessThanOrEqual":    "%s <= %v",
	"inRange":            "%s between %v AND %v",
}

func numberWhere(field string, f filtering.Filter) string {
	if safeDeref(f.Type) == "inRange" {
		return fmt.Sprintf(
			numberOps[safeDeref(f.Type)],
			field,
			safeDeref(f.Filter),
			safeDeref(f.FilterTo),
		)
	} else {
		return fmt.Sprintf(numberOps[safeDeref(f.Type)], field, safeDeref(f.Filter))
	}
}

var dateOps = map[string]string{
	"equals":      "DATE(%s) = '%s'",
	"notEqual":    "DATE(%s) <> '%s'",
	"greaterThan": "DATE(%s) > '%s'",
	"lessThan":    "DATE(%s) < '%s'",
	"inRange":     "DATE(%s) between '%s' AND '%s'",
}

func dateWhere(field string, f filtering.Filter) string {
	typ := safeDeref(f.Type)
	from1, to1 := sliceDate(f.DateFrom), sliceDate(f.DateTo)

	if typ == "inRange" {
		return fmt.Sprintf(dateOps["inRange"], field, from1, to1)
	}

	op := dateOps[typ]

	return fmt.Sprintf(op, field, from1)
}

var textOps = map[string]string{
	"equals":      "lower(%s) = '%s'",
	"notEqual":    "lower(%s) <> '%s'",
	"contains":    "lower(%s) like '%s'",
	"notContains": "lower(%s) not like '%s'",
	"startsWith":  "lower(%s) like '%s'",
	"endsWith":    "lower(%s) like '%s'",
}

func textWhere(field string, f filtering.Filter) string {
	typ := safeDeref(f.Type)
	val := fmt.Sprint(*f.Filter)
	pat := strings.ToLower(val)
	like := map[string]string{
		"equals":      pat,
		"contains":    "%" + pat + "%",
		"notContains": "%" + pat + "%",
		"startsWith":  pat + "%",
		"endsWith":    "%" + pat,
	}[typ]

	return fmt.Sprintf(textOps[typ], field, like)
}

func setWhere(field string, vals []*string) string {
	if len(vals) == 0 {
		return ""
	}

	in := make([]string, 0, len(vals))
	hasNull := false

	for _, v := range vals {
		if v == nil || *v == "" {
			hasNull = true
		} else {
			in = append(in, fmt.Sprintf("'%s'", *v))
		}
	}

	if len(in) == 0 && hasNull {
		return field + " IS NULL"
	}

	joined := strings.Join(in, ",")
	if hasNull {
		return fmt.Sprintf("%s in (%s) OR %s IS NULL", field, joined, field)
	}

	return fmt.Sprintf("%s in (%s)", field, joined)
}

// utility to avoid nil-pointer deref.
func safeDeref[T any](s *T) T {
	if s == nil {
		return *new(T)
	}

	return *s
}

// slice YYYY-MM-DD.
func sliceDate(src *string) string {
	if src == nil || len(*src) < 10 {
		return ""
	}

	return (*src)[:10]
}

// CreateOrder builds SQL ORDER BY clauses from the sort model.
// "prefix" is applied to fields without a dot ("owner.field" remains unchanged).
// Pass empty string for no prefix.
func CreateOrder(sm *filtering.SortModel, prefix string) []string {
	if sm == nil || len(*sm) == 0 {
		return []string{}
	}

	out := make([]string, len(*sm))

	for i, sort := range *sm {
		field := sort["colId"]
		if prefix != "" && !strings.Contains(field, ".") {
			field = prefix + "." + field
		}

		out[i] = fmt.Sprintf("%s %s", field, sort["sort"])
	}

	return out
}
