package jts_test

import (
	"encoding/json"
	"fmt"
	"net/url"
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
			name: "nil value is null",
			args: args{
				in: []any{
					map[string]any{"key": nil},
				},
			},
			want:    `{"key":null}`,
			wantErr: false,
		},
		{
			name: "duplicate nil is retained",
			args: args{
				in: []any{
					jts.Pair("key", "value"),
					map[string]any{"key": nil},
				},
			},
			want:    `{"key":["value",null]}`,
			wantErr: false,
		},
		{
			name: "typed nil slice is null",
			args: args{
				in: []any{
					map[string][]any{"key": nil},
				},
			},
			want:    `{"key":null}`,
			wantErr: false,
		},
		{
			name: "duplicate typed nil slice is retained",
			args: args{
				in: []any{
					jts.Pair("key", "value"),
					map[string][]any{"key": nil},
				},
			},
			want:    `{"key":["value",null]}`,
			wantErr: false,
		},
		{
			name: "typed nil slice before scalar is retained",
			args: args{
				in: []any{
					map[string][]any{"key": nil},
					jts.Pair("key", "value"),
				},
			},
			want:    `{"key":[null,"value"]}`,
			wantErr: false,
		},
		{
			name: "nil before scalar is retained",
			args: args{
				in: []any{
					map[string]any{"key": nil},
					jts.Pair("key", "value"),
				},
			},
			want:    `{"key":[null,"value"]}`,
			wantErr: false,
		},
		{
			name: "nil retains duplicate order",
			args: args{
				in: []any{
					map[string]any{"key": nil},
					jts.Pair("key", "value"),
					"key", "value2",
				},
			},
			want:    `{"key":[null,"value","value2"]}`,
			wantErr: false,
		},
		{
			name: "nil and url values retain duplicate order",
			args: args{
				in: []any{
					map[string]any{"key": nil},
					jts.Pair("key", "value"),
					"key", "value2",
					url.Values{"key": []string{"value3"}},
				},
			},
			want:    `{"key":[null,"value","value2","value3"]}`,
			wantErr: false,
		},
		{
			name: "empty url values",
			args: args{
				in: []any{
					url.Values{},
				},
			},
			want:    `{}`,
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

// FuzzNew exercises scalar, duplicate, nested, nil, map, slice, and byte values.
func FuzzNew(f *testing.F) {
	f.Add([]byte("a"))
	f.Add([]byte("keyvalue"))
	f.Add([]byte{0, 1, 2, 3, 4, 5})

	f.Fuzz(func(t *testing.T, data []byte) {
		var inputs []any
		for i, b := range data {
			if i%2 == 0 {
				inputs = append(inputs, string([]byte{b}))
				continue
			}

			switch b % 7 {
			case 0:
				inputs = append(inputs, nil)
			case 1:
				inputs = append(inputs, []int{int(b), int(b + 1)})
			case 2:
				inputs = append(inputs, map[string]any{"byte": b, "nil": nil})
			case 3:
				inputs = append(inputs, [][]int{{int(b)}, {int(b + 1)}})
			case 4:
				inputs = append(inputs, []byte{b})
			case 5:
				inputs = append(inputs, struct {
					Value byte `json:"value"`
				}{Value: b})
			default:
				inputs = append(inputs, int(b))
			}
		}

		callArgs := []any{[]any{inputs}, "tail", map[string]any{"nested": data}}
		wantErr := len(data) > 0 && data[0]%11 == 0
		if wantErr {
			callArgs = append(callArgs, jts.Pair("marshal_error", failingMarshaler{}))
		}

		result, err := jts.New(callArgs...)
		if wantErr {
			if err == nil {
				t.Fatalf("New returned nil error for a failing custom marshaler")
			}
			if result != "" {
				t.Fatalf("New returned %q with a marshaling error, want empty string", result)
			}
			return
		}
		if err != nil {
			t.Fatalf("New returned error: %v", err)
		}

		var m map[string]any
		if err := json.Unmarshal([]byte(result), &m); err != nil {
			t.Fatalf("Result is not valid JSON: %v", err)
		}
	})
}

func ExampleNew() {
	result, err := jts.New("key", "value", "jump", true)
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println(result)
	// Output:
	// {"jump":true,"key":"value"}
}

func ExampleNew_pair() {
	result, err := jts.New(jts.Pair("key", "value"))
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println(result)
	// Output:
	// {"key":"value"}
}

func ExampleNew_mapWithSlice() {
	result, err := jts.New(map[string][]string{"names": {"john", "jane"}})
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println(result)
	// Output:
	// {"names":["john","jane"]}
}

func ExampleNew_map() {
	result, err := jts.New(map[string]string{"name": "john"})
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println(result)
	// Output:
	// {"name":"john"}
}

func ExampleNew_mix() {
	result, err := jts.New(map[string][]string{"names": {"john", "jane"}}, jts.Pair("key", "value"), "key", "value2")
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println(result)
	// Output:
	// {"key":["value","value2"],"names":["john","jane"]}
}

func ExampleNew_multipleKeys() {
	result, err := jts.New("key", "value", "jump", true, "key", "value2")
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println(result)
	// Output:
	// {"jump":true,"key":["value","value2"]}
}

func ExampleNew_singleValue() {
	result, err := jts.New("value")
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println(result)
	// Output:
	// {"!INVALID":"value"}
}
