package raw_filtering

import (
	"reflect"
	"testing"

	"github.com/SOTBI-LLC/sotbi.lib/pkg/filtering"
)

func strptr(s string) *string { return &s }

func any2ptr[T any](s T) *any {
	v := any(s)
	return &v
}

func TestSetWhere(t *testing.T) {
	cases := []struct {
		name   string
		vals   []*string
		expect string
	}{
		{"only values", []*string{strptr("a"), strptr("b")}, "f in ('a','b')"},
		{"with null", []*string{nil, strptr("x")}, "f in ('x') OR f IS NULL"},
		{"only null", []*string{nil, strptr("")}, "f IS NULL"},
		{"empty slice", []*string{}, ""},
	}
	for _, c := range cases {
		got := setWhere("f", c.vals)
		if got != c.expect {
			t.Errorf("%s: got %q; want %q", c.name, got, c.expect)
		}
	}
}

func TestDateWhere(t *testing.T) {
	df := strptr("2025-05-23T00:00:00Z")
	dt := strptr("2025-05-30T23:59:59Z")

	cases := []struct {
		name string
		f    filtering.Filter
		want string
	}{
		{
			"single equals",
			filtering.Filter{Type: strptr("equals"), DateFrom: df},
			"DATE(f) = '2025-05-23'",
		},
		{
			"inRange",
			filtering.Filter{Type: strptr("inRange"), DateFrom: df, DateTo: dt},
			"DATE(f) between '2025-05-23' AND '2025-05-30'",
		},
	}
	for _, c := range cases {
		if got := dateWhere("f", c.f); got != c.want {
			t.Errorf("%s: got %q; want %q", c.name, got, c.want)
		}
	}
}

func TestNumberWhere(t *testing.T) {
	n := 42
	n2 := "100"

	cases := []struct {
		name string
		f    filtering.Filter
		want string
	}{
		{"equals", filtering.Filter{Type: strptr("equals"), Filter: any2ptr(n)}, "f = 42"},
		{
			"inRange",
			filtering.Filter{Type: strptr("inRange"), Filter: any2ptr(n), FilterTo: any2ptr(n2)},
			"f between 42 AND 100",
		},
	}
	for _, c := range cases {
		if got := numberWhere("f", c.f); got != c.want {
			t.Errorf("%s: got %q; want %q", c.name, got, c.want)
		}
	}
}

func TestTextWhere(t *testing.T) {
	cases := []struct {
		name string
		typ  string
		val  string
		want string
	}{
		{"equals", "equals", "AbC", "lower(f) = 'abc'"},
		{"contains", "contains", "Xy", "lower(f) like '%xy%'"},
		{"startsWith", "startsWith", "Hey", "lower(f) like 'hey%'"},
		{"endsWith", "endsWith", "Lo", "lower(f) like '%lo'"},
	}
	for _, c := range cases {
		val := any(c.val)

		f := filtering.Filter{Type: strptr(c.typ), Filter: &val}
		if got := textWhere("f", f); got != c.want {
			t.Errorf("%s: got %q; want %q", c.name, got, c.want)
		}
	}
}

func TestCreateFilter(t *testing.T) {
	m := filtering.FilterModel{
		"a": {FilterType: strptr("set"), Values: []*string{strptr("x")}},
		"b": {FilterType: strptr("text"), Type: strptr("contains"), Filter: any2ptr("Z")},
	}
	out := CreateFilter(m, strptr("pre"))
	want1 := "pre.a in ('x')"
	want2 := "lower(pre.b) like '%z%'"

	if len(out) != 2 || (out[0] != want1 && out[1] != want1) {
		t.Errorf("CreateFilter missing set; got %v", out)
	}

	if len(out) != 2 || (out[0] != want2 && out[1] != want2) {
		t.Errorf("CreateFilter missing text; got %v", out)
	}
}

func TestSortModel_Unmarshal(t *testing.T) {
	jsonData := []byte(
		`[
		  {"colId":"name","sort":"asc"},
		  {"colId":"owner.field","sort":"desc"}
		]`,
	)

	var sm filtering.SortModel

	if err := sm.Unmarshal(jsonData); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if len(sm) != 2 {
		t.Fatalf("Expected 2 items, got %d", len(sm))
	}

	if sm[0]["colId"] != "name" || sm[0]["sort"] != "asc" {
		t.Errorf("Unexpected first element: %v", sm[0])
	}

	if sm[1]["colId"] != "owner.field" || sm[1]["sort"] != "desc" {
		t.Errorf("Unexpected second element: %v", sm[1])
	}
}

func TestSortModel_Unmarshal_InvalidJSON(t *testing.T) {
	var sm filtering.SortModel

	err := sm.Unmarshal([]byte(`invalid`))
	if err == nil {
		t.Errorf("Expected error for invalid JSON, got nil")
	}
}

func TestSortModel_CreateOrder(t *testing.T) {
	sm := filtering.SortModel{
		{"colId": "name", "sort": "ASC"},
		{"colId": "owner.field", "sort": "DESC"},
		{"colId": "age", "sort": "desc"},
	}

	cases := []struct {
		prefix string
		want   []string
	}{
		{
			prefix: "",
			want: []string{
				"name ASC",
				"owner.field DESC",
				"age desc",
			},
		},
		{
			prefix: "u",
			want: []string{
				"u.name ASC",
				"owner.field DESC",
				"u.age desc",
			},
		},
	}

	for _, tc := range cases {
		got := CreateOrder(&sm, tc.prefix)
		if !reflect.DeepEqual(got, tc.want) {
			t.Errorf("CreateOrder(%q) = %v; want %v", tc.prefix, got, tc.want)
		}
	}
}
