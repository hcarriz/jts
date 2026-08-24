package jts_test

import (
	"strconv"
	"testing"

	"github.com/hcarriz/jts"
)

var benchmarkOutput string

type benchmarkPayload struct {
	Name    string         `json:"name"`
	Enabled bool           `json:"enabled"`
	Scores  []int          `json:"scores"`
	Labels  map[string]any `json:"labels"`
}

func BenchmarkNew(b *testing.B) {
	large := make([]any, 0, 256)
	for i := range 128 {
		large = append(large,
			"field_"+strconv.Itoa(i),
			benchmarkPayload{
				Name:    "value_" + strconv.Itoa(i),
				Enabled: i%2 == 0,
				Scores:  []int{i, i + 1, i + 2},
				Labels: map[string]any{
					"group": i % 8,
					"valid": true,
				},
			},
		)
	}

	tests := []struct {
		name string
		in   []any
	}{
		{
			name: "small",
			in:   []any{"method", "POST", "status", 200, "cached", false},
		},
		{
			name: "duplicate_slices",
			in: []any{
				"tags", []string{"go", "json"},
				jts.Pair("tags", [2]string{"security", "htmx"}),
			},
		},
		{
			name: "large_nested",
			in:   large,
		},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			b.ReportAllocs()

			var result string
			for b.Loop() {
				var err error
				result, err = jts.New(tt.in...)
				if err != nil {
					b.Fatalf("New() error = %v", err)
				}
			}

			benchmarkOutput = result
			b.ReportMetric(float64(len(result)), "output-B/op")
		})
	}
}
