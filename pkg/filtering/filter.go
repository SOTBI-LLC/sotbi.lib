// repo/filter.go
package filtering

import (
	"encoding/json"
	"fmt"
	"strings"
)

type Condition struct {
	Filter
}

type Filter struct {
	Type       *string    `json:"type,omitempty"`
	FilterType *string    `json:"filterType,omitempty"`
	Operator   *string    `json:"operator,omitempty"`
	Values     *[]*string `json:"values,omitempty"`
	DateFrom   *string    `json:"dateFrom,omitempty"`
	DateTo     *string    `json:"dateTo,omitempty"`
	Filter     *any       `json:"filter,omitempty"`
	FilterTo   *any       `json:"filterTo,omitempty"`
	Condition1 *Condition `json:"condition1,omitempty"`
	Condition2 *Condition `json:"condition2,omitempty"`
}

func (f Filter) IsEmpty() bool {
	return f.FilterType == nil &&
		f.Operator == nil &&
		f.Values == nil &&
		f.DateFrom == nil && f.DateTo == nil &&
		f.Filter == nil && f.FilterTo == nil &&
		f.Condition1 == nil && f.Condition2 == nil
}

type FilterModel map[string]Filter

func (fm FilterModel) Unmarshal(data []byte) error {
	err := json.Unmarshal(data, &fm)
	if err != nil {
		return err
	}

	return nil
}

// SortModel model.
type SortModel []map[string]string

func (sm *SortModel) Unmarshal(data []byte) error {
	return json.Unmarshal(data, &sm)
}

// CreateOrder builds SQL ORDER BY clauses from the sort model.
// "prefix" is applied to fields without a dot ("owner.field" remains unchanged).
// Pass empty string for no prefix.
func (sm *SortModel) CreateOrder(prefix string) []string {
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

// ParseJSONToFilterModel constructor.
func ParseJSONToFilterModel(args string) (filterModel FilterModel, err error) {
	err = json.Unmarshal([]byte(args), &filterModel)
	if err != nil {
		return filterModel, err
	}

	return filterModel, err
}
