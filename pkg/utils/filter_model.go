package utils

import "encoding/json"

// SortModel model.
type SortModel []map[string]string

// Filter model.
type Filter struct {
	Type       *string    `json:"type,omitempty"`
	FilterType *string    `json:"filterType,omitempty"`
	Operator   *string    `json:"operator,omitempty"`
	Values     *[]string  `json:"values,omitempty"`
	DateFrom   *string    `json:"dateFrom,omitempty"`
	DateTo     *string    `json:"dateTo,omitempty"`
	Filter     *any       `json:"filter,omitempty"`
	FilterTo   *any       `json:"filterTo,omitempty"`
	Condition1 *Condition `json:"condition1,omitempty"`
	Condition2 *Condition `json:"condition2,omitempty"`
}

// FilterModel model.
type FilterModel map[string]Filter

// Condition struct.
type Condition struct {
	Filter
}

// ParseJSONToFilterModel constructor.
func ParseJSONToFilterModel(args string) (filterModel FilterModel, err error) {
	err = json.Unmarshal([]byte(args), &filterModel)
	if err != nil {
		return filterModel, err
	}

	return filterModel, err
}
