package jts

import (
	"encoding"
	"encoding/json"
	"net/url"
	"reflect"
)

// BadKey is the key used for a trailing string or an unsupported standalone
// input. It is an ordinary key, so generated and explicit BadKey values combine
// using the same duplicate-key rules as any other key.
const BadKey = "!INVALID"

// Input is an opaque key/value pair consumed by [Convert] and [New]. Construct
// values with [Pair]. The zero value is valid and is ignored.
type Input struct {
	key   string
	value any
}

// Pair returns an Input containing key and value. An Input with an empty key is
// consumed but omitted by [Convert] and [New]. Values retain their encoding/json
// representation unless duplicate-key array concatenation combines them.
func Pair[T any](key string, value T) Input {
	return Input{
		key:   key,
		value: value,
	}
}

type sliceID struct {
	data uintptr
	len  int
	cap  int
}

type argumentFrame struct {
	args    []any
	index   int
	id      sliceID
	tracked bool
}

type argumentReader struct {
	frames []argumentFrame
	active map[sliceID]struct{}
}

func newArgumentReader(args []any) *argumentReader {
	r := &argumentReader{active: make(map[sliceID]struct{})}
	r.push(args, true)
	return r
}

func (r *argumentReader) nextInput() (Input, bool) {
	for {
		value, ok := r.next()
		if !ok {
			return Input{}, false
		}

		switch value := value.(type) {
		case string:
			paired, ok := r.next()
			if !ok {
				return Pair(BadKey, value), true
			}
			return Pair(value, paired), true
		case Input:
			return value, true
		case map[string][]any:
			r.push(mapInputs(value), false)
		case map[string]any:
			r.push(mapInputs(value), false)
		case map[string]string:
			r.push(mapInputs(value), false)
		case map[string][]string:
			r.push(mapInputs(value), false)
		case []any:
			if !r.push(value, true) {
				return Pair(BadKey, value), true
			}
		case url.Values:
			r.push(mapInputs(value), false)
		default:
			return Pair(BadKey, value), true
		}
	}
}

func (r *argumentReader) next() (any, bool) {
	for len(r.frames) > 0 {
		last := len(r.frames) - 1
		frame := &r.frames[last]
		if frame.index < len(frame.args) {
			value := frame.args[frame.index]
			frame.index++
			return value, true
		}

		if frame.tracked {
			delete(r.active, frame.id)
		}
		r.frames = r.frames[:last]
	}

	return nil, false
}

func (r *argumentReader) push(args []any, track bool) bool {
	if len(args) == 0 {
		return true
	}

	frame := argumentFrame{args: args}
	if track {
		frame.id = identifySlice(args)
		if _, exists := r.active[frame.id]; exists {
			return false
		}
		frame.tracked = true
		r.active[frame.id] = struct{}{}
	}

	r.frames = append(r.frames, frame)
	return true
}

func identifySlice(value []any) sliceID {
	return sliceID{
		data: reflect.ValueOf(value).Pointer(),
		len:  len(value),
		cap:  cap(value),
	}
}

func mapInputs[T any](values map[string]T) []any {
	inputs := make([]any, 0, len(values))
	for key, value := range values {
		if key != "" {
			inputs = append(inputs, Pair(key, value))
		}
	}
	return inputs
}

// Convert processes in as the compatibility input forms described in the
// package documentation and returns a non-nil map. Duplicate values are combined
// in encounter order. Arrays concatenate when duplicate keys are combined;
// supplied maps and structs remain values and are never deep-merged.
//
// Convert does not validate that values can be encoded as JSON. Use [New] when a
// JSON string and encoding errors are required. The returned map has unspecified
// iteration order.
func Convert(in ...any) map[string]any {
	list := make(map[string]any)
	reader := newArgumentReader(in)

	for {
		input, ok := reader.nextInput()
		if !ok {
			return list
		}
		if input.key == "" {
			continue
		}

		addValue(list, input.key, input.value)
	}
}

func addValue(values map[string]any, key string, value any) {
	existing, duplicate := values[key]
	if !duplicate {
		values[key] = value
		return
	}

	combined := appendDuplicateValue(nil, existing)
	values[key] = appendDuplicateValue(combined, value)
}

func appendDuplicateValue(values []any, value any) []any {
	if !concatenatesOnDuplicate(value) {
		return append(values, value)
	}

	reflected := reflect.ValueOf(value)
	for i := range reflected.Len() {
		values = append(values, reflected.Index(i).Interface())
	}
	return values
}

func concatenatesOnDuplicate(value any) bool {
	if value == nil {
		return false
	}
	if _, ok := value.(json.Marshaler); ok {
		return false
	}
	if _, ok := value.(encoding.TextMarshaler); ok {
		return false
	}

	reflected := reflect.ValueOf(value)
	switch reflected.Kind() {
	case reflect.Array:
		return true
	case reflect.Slice:
		if reflected.IsNil() || reflected.Type().Elem().Kind() == reflect.Uint8 {
			return false
		}
		return true
	default:
		return false
	}
}

// New converts in with [Convert] and marshals the result as compact JSON. It is
// the primary API for producing a templ dynamic value for an htmx hx-vals
// attribute.
//
// If encoding/json cannot marshal a value, New returns an empty string and that
// error. New does not mark its result as trusted HTML or JavaScript; callers must
// preserve templ's normal dynamic-attribute escaping.
func New(in ...any) (string, error) {
	result, err := json.Marshal(Convert(in...))
	if err != nil {
		return "", err
	}

	return string(result), nil
}
