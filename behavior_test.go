package jts_test

import (
	"encoding/json"
	"errors"
	"maps"
	"math"
	"net/url"
	"slices"
	"testing"

	"github.com/hcarriz/jts"
)

type namedString string
type namedMap map[string]string
type namedSlice []int

var errMarshal = errors.New("marshal failure")

type failingMarshaler struct{}

func (failingMarshaler) MarshalJSON() ([]byte, error) {
	return nil, errMarshal
}

type objectMarshaler []int

func (objectMarshaler) MarshalJSON() ([]byte, error) {
	return []byte(`{"custom":true}`), nil
}

type textMarshaledSlice []string

func (value textMarshaledSlice) MarshalText() ([]byte, error) {
	return []byte(value[0]), nil
}

func TestCurrentInputForms(t *testing.T) {
	tests := []struct {
		name string
		in   []any
		want string
	}{
		{
			name: "no input",
			want: `{}`,
		},
		{
			name: "string pair",
			in:   []any{"key", "value"},
			want: `{"key":"value"}`,
		},
		{
			name: "Input from Pair",
			in:   []any{jts.Pair("key", "value")},
			want: `{"key":"value"}`,
		},
		{
			name: "map string to any",
			in:   []any{map[string]any{"key": 1}},
			want: `{"key":1}`,
		},
		{
			name: "map string to string",
			in:   []any{map[string]string{"key": "value"}},
			want: `{"key":"value"}`,
		},
		{
			name: "map string to any slice",
			in:   []any{map[string][]any{"key": {1, "two"}}},
			want: `{"key":[1,"two"]}`,
		},
		{
			name: "map string to string slice",
			in:   []any{map[string][]string{"key": {"one", "two"}}},
			want: `{"key":["one","two"]}`,
		},
		{
			name: "url Values",
			in:   []any{url.Values{"key": {"one", "two"}}},
			want: `{"key":["one","two"]}`,
		},
		{
			name: "any slice argument list",
			in:   []any{[]any{"one", 1, "two", 2}},
			want: `{"one":1,"two":2}`,
		},
		{
			name: "string consumes supported form as value",
			in:   []any{"key", map[string]string{"nested": "value"}},
			want: `{"key":{"nested":"value"}}`,
		},
		{
			name: "string consumes any slice as an array value",
			in:   []any{"key", []any{"nested", 2}},
			want: `{"key":["nested",2]}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := jts.New(tt.in...)
			if err != nil {
				t.Fatalf("New() error = %v", err)
			}
			if got != tt.want {
				t.Errorf("New() = %s, want %s", got, tt.want)
			}
		})
	}
}

func TestNewRoundTripsNestedValues(t *testing.T) {
	type filter struct {
		Status string         `json:"status"`
		Tags   []string       `json:"tags"`
		Scores map[string]int `json:"scores"`
	}
	type payload struct {
		Filter  filter           `json:"filter"`
		IDs     []int            `json:"ids"`
		Matrix  [][]int          `json:"matrix"`
		Blob    []byte           `json:"blob"`
		Raw     map[string]bool  `json:"raw"`
		Objects []map[string]int `json:"objects"`
		Nothing *int             `json:"nothing"`
	}

	wantFilter := filter{
		Status: "open",
		Tags:   []string{"go", "security"},
		Scores: map[string]int{"quality": 10},
	}
	encoded, err := jts.New(
		"filter", wantFilter,
		"ids", []int{1, 2},
		jts.Pair("ids", [2]int{3, 4}),
		"matrix", [][]int{{1, 2}, {3, 4}},
		"blob", []byte("AB"),
		"raw", json.RawMessage(`{"enabled":true}`),
		"objects", map[string]int{"one": 1},
		"objects", map[string]int{"two": 2},
		"nothing", nil,
	)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	var got payload
	if err := json.Unmarshal([]byte(encoded), &got); err != nil {
		t.Fatalf("unmarshal generated JSON: %v", err)
	}

	if got.Filter.Status != wantFilter.Status ||
		!slices.Equal(got.Filter.Tags, wantFilter.Tags) ||
		!maps.Equal(got.Filter.Scores, wantFilter.Scores) {
		t.Errorf("Filter = %#v, want %#v", got.Filter, wantFilter)
	}
	if !slices.Equal(got.IDs, []int{1, 2, 3, 4}) {
		t.Errorf("IDs = %v, want [1 2 3 4]", got.IDs)
	}
	if len(got.Matrix) != 2 ||
		!slices.Equal(got.Matrix[0], []int{1, 2}) ||
		!slices.Equal(got.Matrix[1], []int{3, 4}) {
		t.Errorf("Matrix = %v, want [[1 2] [3 4]]", got.Matrix)
	}
	if !slices.Equal(got.Blob, []byte("AB")) {
		t.Errorf("Blob = %q, want %q", got.Blob, []byte("AB"))
	}
	if !maps.Equal(got.Raw, map[string]bool{"enabled": true}) {
		t.Errorf("Raw = %v, want map[enabled:true]", got.Raw)
	}
	if len(got.Objects) != 2 ||
		!maps.Equal(got.Objects[0], map[string]int{"one": 1}) ||
		!maps.Equal(got.Objects[1], map[string]int{"two": 2}) {
		t.Errorf("Objects = %v, want [map[one:1] map[two:2]]", got.Objects)
	}
	if got.Nothing != nil {
		t.Errorf("Nothing = %v, want nil", *got.Nothing)
	}

	var fields map[string]json.RawMessage
	if err := json.Unmarshal([]byte(encoded), &fields); err != nil {
		t.Fatalf("unmarshal generated JSON fields: %v", err)
	}
	if got := string(fields["nothing"]); got != "null" {
		t.Errorf("nothing JSON = %s, want null", got)
	}
}

func TestArgumentListsAreCycleSafe(t *testing.T) {
	cycle := make([]any, 1)
	cycle[0] = cycle

	converted := jts.Convert(cycle)
	if _, ok := converted[jts.BadKey]; !ok {
		t.Fatal("Convert(cyclic argument list) did not retain the cycle under BadKey")
	}

	got, err := jts.New(cycle)
	if err == nil {
		t.Fatalf("New(cyclic argument list) = %q, nil error; want a JSON cycle error", got)
	}
	if got != "" {
		t.Errorf("New(cyclic argument list) result = %q, want empty string", got)
	}
}

func TestDeepArgumentListsDoNotRequireRecursiveParsing(t *testing.T) {
	args := []any{"key", "value"}
	for range 10_000 {
		args = []any{args}
	}

	got, err := jts.New(args)
	if err != nil {
		t.Fatalf("New(deep argument list) error = %v", err)
	}
	if got != `{"key":"value"}` {
		t.Errorf("New(deep argument list) = %s, want %s", got, `{"key":"value"}`)
	}
}

func TestConversionBehavior(t *testing.T) {
	tests := []struct {
		name string
		in   []any
		want string
	}{
		{
			name: "trailing string uses BadKey",
			in:   []any{"value"},
			want: `{"!INVALID":"value"}`,
		},
		{
			name: "named string uses BadKey",
			in:   []any{namedString("value")},
			want: `{"!INVALID":"value"}`,
		},
		{
			name: "named map uses BadKey",
			in:   []any{namedMap{"key": "value"}},
			want: `{"!INVALID":{"key":"value"}}`,
		},
		{
			name: "named slice uses BadKey",
			in:   []any{namedSlice{1, 2}},
			want: `{"!INVALID":[1,2]}`,
		},
		{
			name: "nested argument list preserves outer remainder",
			in:   []any{"kept", 1, []any{"nested", 2}, "discarded", 3},
			want: `{"discarded":3,"kept":1,"nested":2}`,
		},
		{
			name: "empty argument list preserves outer remainder",
			in:   []any{[]any{}, "discarded", 1},
			want: `{"discarded":1}`,
		},
		{
			name: "nil argument list preserves outer remainder",
			in:   []any{[]any(nil), "discarded", 1},
			want: `{"discarded":1}`,
		},
		{
			name: "argument list is expanded in place",
			in:   []any{[]any{"key"}, "value"},
			want: `{"key":"value"}`,
		},
		{
			name: "empty keys and zero Input are omitted",
			in: []any{
				jts.Pair("", 1),
				"", "value",
				map[string]string{"": "ignored", "kept": "value"},
				jts.Input{},
			},
			want: `{"kept":"value"}`,
		},
		{
			name: "lone empty string uses BadKey",
			in:   []any{""},
			want: `{"!INVALID":""}`,
		},
		{
			name: "explicit and generated BadKey values aggregate",
			in:   []any{jts.Pair(jts.BadKey, "explicit"), true, 7},
			want: `{"!INVALID":["explicit",true,7]}`,
		},
		{
			name: "duplicates and arrays flatten in encounter order",
			in: []any{
				"key", "first",
				jts.Pair("key", []string{"second", "third"}),
				jts.Pair("key", [1]int{4}),
				jts.Pair("key", []any{[]int{5, 6}}),
			},
			want: `{"key":["first","second","third",4,[5,6]]}`,
		},
		{
			name: "empty duplicate slice converts scalar to array",
			in:   []any{"key", "value", jts.Pair("key", []string{})},
			want: `{"key":["value"]}`,
		},
		{
			name: "byte slice retains encoding/json representation",
			in:   []any{jts.Pair("key", []byte{'A', 'B'})},
			want: `{"key":"QUI="}`,
		},
		{
			name: "RawMessage retains custom JSON encoding",
			in:   []any{jts.Pair("key", json.RawMessage(`{"raw":true}`))},
			want: `{"key":{"raw":true}}`,
		},
		{
			name: "JSON marshalers remain atomic across duplicates",
			in: []any{
				jts.Pair("key", objectMarshaler{1, 2}),
				jts.Pair("key", objectMarshaler{3, 4}),
			},
			want: `{"key":[{"custom":true},{"custom":true}]}`,
		},
		{
			name: "text marshalers remain atomic across duplicates",
			in: []any{
				jts.Pair("key", textMarshaledSlice{"first"}),
				jts.Pair("key", textMarshaledSlice{"second"}),
			},
			want: `{"key":["first","second"]}`,
		},
		{
			name: "first untyped nil becomes null",
			in:   []any{jts.Pair[any]("key", nil)},
			want: `{"key":null}`,
		},
		{
			name: "later untyped nil is retained",
			in:   []any{"key", "value", jts.Pair[any]("key", nil)},
			want: `{"key":["value",null]}`,
		},
		{
			name: "first nil followed by value aggregates null",
			in:   []any{jts.Pair[any]("key", nil), jts.Pair("key", "value")},
			want: `{"key":[null,"value"]}`,
		},
		{
			name: "typed nil values retain type behavior",
			in: []any{
				jts.Pair("slice", []string(nil)),
				jts.Pair("pointer", (*int)(nil)),
				jts.Pair("map", map[string]int(nil)),
				map[string][]any(nil),
				url.Values(nil),
			},
			want: `{"map":null,"pointer":null,"slice":null}`,
		},
		{
			name: "JSON output is compact and sorts object keys",
			in:   []any{"z", 1, "a", 2},
			want: `{"a":2,"z":1}`,
		},
		{
			name: "JSON escapes HTML and line separator characters",
			in:   []any{"key", "<>&\u2028\u2029\"\\"},
			want: `{"key":"\u003c\u003e\u0026\u2028\u2029\"\\"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := jts.New(tt.in...)
			if err != nil {
				t.Fatalf("New() error = %v", err)
			}
			if got != tt.want {
				t.Errorf("New() = %s, want %s", got, tt.want)
			}
		})
	}
}

func TestConvertReturnsNonNilMapWithoutJSONValidation(t *testing.T) {
	got := jts.Convert("unsupported", func() {})
	if got == nil {
		t.Fatal("Convert() returned a nil map")
	}
	if _, ok := got["unsupported"]; !ok {
		t.Fatal("Convert() omitted a value unsupported by encoding/json")
	}
}

func TestNewMarshalErrors(t *testing.T) {
	cycle := map[string]any{}
	cycle["self"] = cycle

	tests := []struct {
		name  string
		value any
	}{
		{name: "function", value: func() {}},
		{name: "channel", value: make(chan int)},
		{name: "complex", value: complex(1, 2)},
		{name: "NaN", value: math.NaN()},
		{name: "positive infinity", value: math.Inf(1)},
		{name: "cycle", value: cycle},
		{name: "custom marshaler", value: failingMarshaler{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := jts.New("key", tt.value)
			if err == nil {
				t.Fatalf("New() = %q, nil error; want an encoding error", got)
			}
			if got != "" {
				t.Errorf("New() result = %q, want empty string", got)
			}
		})
	}

	_, err := jts.New("key", failingMarshaler{})
	if !errors.Is(err, errMarshal) {
		t.Errorf("New() error = %v, want error wrapping %v", err, errMarshal)
	}
}
