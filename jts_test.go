package jts_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/hcarriz/jts"
)

func TestNew(t *testing.T) {
	type args struct {
		in []any
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name:    "empty",
			args:    args{},
			want:    "{}",
			wantErr: false,
		},
		{
			name: "strings",
			args: args{
				in: []any{
					"key", "value",
				},
			},
			want:    `{"key":"value"}`,
			wantErr: false,
		},
		{
			name: "odd strings",
			args: args{
				in: []any{
					"key", "value", "bad",
				},
			},
			want:    `{"!INVALID":"bad","key":"value"}`,
			wantErr: false,
		},
		{
			name: "inputs",
			args: args{
				in: []any{
					jts.Pair("key", "value"),
				},
			},
			want:    `{"key":"value"}`,
			wantErr: false,
		},
		{
			name: "duplicate inputs",
			args: args{
				in: []any{
					jts.Pair("key", "value"), jts.Pair("key", "value2"),
				},
			},
			want:    `{"key":["value","value2"]}`,
			wantErr: false,
		},
		{
			name: "map",
			args: args{
				in: []any{
					map[string]string{
						"name": "john doe",
					},
				},
			},
			want:    `{"name":"john doe"}`,
			wantErr: false,
		},
		{
			name: "map with array",
			args: args{
				in: []any{
					map[string][]string{
						"name": {"john", "sally"},
					},
				},
			},
			want:    `{"name":["john","sally"]}`,
			wantErr: false,
		},
		{
			name: "mix",
			args: args{
				in: []any{
					jts.Pair("key", "value"),
					"key2", "value2",
					map[string][]string{
						"key": {"one", "two"},
						"map": {"value"},
					},
				},
			},
			want:    `{"key":["value","one","two"],"key2":"value2","map":["value"]}`,
			wantErr: false,
		},
		{
			name: "integer map",
			args: args{
				in: []any{
					map[string]any{"key": 1},
				},
			},
			want:    `{"key":1}`,
			wantErr: false,
		},
		{
			name: "nil map",
			args: args{
				in: []any{
					map[string]any{},
				},
			},
			want:    `{}`,
			wantErr: false,
		},
		{
			name: "nil value",
			args: args{
				in: []any{
					map[string]any{"key": nil},
				},
			},
			want:    `{"key":""}`,
			wantErr: false,
		},
		{
			name: "nil value 2",
			args: args{
				in: []any{
					jts.Pair("key", "value"),
					map[string]any{"key": nil},
				},
			},
			want:    `{"key":"value"}`,
			wantErr: false,
		},
		{
			name: "nil value 3",
			args: args{
				in: []any{
					map[string][]any{"key": nil},
				},
			},
			want:    `{"key":[]}`,
			wantErr: false,
		},
		{
			name: "nil value 4",
			args: args{
				in: []any{
					jts.Pair("key", "value"),
					map[string][]any{"key": nil},
				},
			},
			want:    `{"key":["value"]}`,
			wantErr: false,
		},
		{
			name: "nil value 5",
			args: args{
				in: []any{
					map[string][]any{"key": nil},
					jts.Pair("key", "value"),
				},
			},
			want:    `{"key":["value"]}`,
			wantErr: false,
		},
		{
			name: "nil value 6",
			args: args{
				in: []any{
					map[string]any{"key": nil},
					jts.Pair("key", "value"),
				},
			},
			want:    `{"key":["","value"]}`,
			wantErr: false,
		},
		{
			name: "nil value 7",
			args: args{
				in: []any{
					map[string]any{"key": nil},
					jts.Pair("key", "value"),
					"key", "value2",
				},
			},
			want:    `{"key":["","value","value2"]}`,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := jts.New(tt.args.in...)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("New() = %v, want %v", got, tt.want)
			}
		})
	}
}

// FuzzNew fuzzes the New function by converting random byte slices into a set of inputs.
func FuzzNew(f *testing.F) {
	// Seed corpus: various byte slices to start with.
	f.Add([]byte("a"))
	f.Add([]byte("keyvalue"))
	f.Add([]byte{0, 1, 2, 3, 4, 5})

	f.Fuzz(func(t *testing.T, data []byte) {
		// Build a slice of inputs.
		// Even indices are used as keys (as one-character strings).
		// Odd indices are used as values; even bytes become an array, odd bytes a simple integer.
		var inputs []any
		for i, b := range data {
			if i%2 == 0 {
				// Use the byte as a key (a string). Note: empty string keys will be skipped.
				inputs = append(inputs, string([]byte{b}))
			} else {
				// For the value, sometimes create a slice (if b is even) or a single integer.
				if b%2 == 0 {
					inputs = append(inputs, []int{int(b), int(b + 1)})
				} else {
					inputs = append(inputs, int(b))
				}
			}
		}

		// Call the function under test.
		result, err := jts.New(inputs...)
		if err != nil {
			t.Fatalf("New returned error: %v", err)
		}

		// Verify the result is valid JSON.
		var m map[string]any
		if err := json.Unmarshal([]byte(result), &m); err != nil {
			t.Fatalf("Result is not valid JSON: %v", err)
		}
	})
}

func ExampleNew() {

	result, _ := jts.New("key", "value", "jump", true)
	fmt.Println(result)
	// Output:
	// {"jump":true,"key":"value"}

}

func ExampleNew_pair() {
	result, _ := jts.New(jts.Pair("key", "value"))
	fmt.Println(result)
	// Output:
	// {"key":"value"}
}

func ExampleNew_mapWithSlice() {
	result, _ := jts.New(map[string][]string{"names": {"john", "jane"}})
	fmt.Println(result)
	// Output:
	// {"names":["john","jane"]}
}

func ExampleNew_map() {
	result, _ := jts.New(map[string]string{"name": "john"})
	fmt.Println(result)
	// Output:
	// {"name":"john"}
}

func ExampleNew_mix() {
	result, _ := jts.New(map[string][]string{"names": {"john", "jane"}}, jts.Pair("key", "value"), "key", "value2")
	fmt.Println(result)
	// Output:
	// {"key":["value","value2"],"names":["john","jane"]}
}

func ExampleNew_multipleKeys() {

	result, _ := jts.New("key", "value", "jump", true, "key", "value2")
	fmt.Println(result)
	// Output:
	// {"jump":true,"key":["value","value2"]}

}

func ExampleNew_singleValue() {

	result, _ := jts.New("value")
	fmt.Println(result)
	// Output:
	// {"!INVALID":"value"}

}
