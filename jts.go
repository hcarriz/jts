package jts

import (
	"encoding/json"
	"reflect"
)

const BadKey = "!INVALID"

type Input struct {
	key   string
	value any
}

func Pair[T any](key string, value T) Input {
	return Input{
		key:   key,
		value: value,
	}
}

func argsToInput(args []any) (Input, []any) {
	if len(args) == 0 {
		return Input{}, nil
	}
	switch x := args[0].(type) {
	case string:
		if len(args) == 1 {
			return Pair(BadKey, x), nil
		}
		return Pair(x, args[1]), args[2:]
	case Input:
		return x, args[1:]
	case map[string][]any:
		return mapToInput(x, args)
	case map[string]any:
		return mapToInput(x, args)
	case map[string]string:
		return mapToInput(x, args)
	case map[string][]string:
		return mapToInput(x, args)
	case []any:
		return argsToInput(x)
	default:
		return Pair(BadKey, x), args[1:]
	}
}

func mapToInput[T any](in map[string]T, args []any) (Input, []any) {
	var list []any
	for key, value := range in {
		if key != "" {
			list = append(list, Pair(key, value))
		}
	}
	if x := args[1:]; len(x) > 0 {
		return Input{}, append(list, x)
	} else {
		return Input{}, list

	}
}

func Convert(in ...any) map[string]any {

	list := make(map[string]any)

	var a Input

	for len(in) > 0 {
		a, in = argsToInput(in)
		if a.key == "" {
			continue
		}

		// Check if the key already exists in the map.
		if existing, ok := list[a.key]; ok {
			var slice []any
			// Convert existing value into a slice if it isn't already one.
			existingKind := reflect.TypeOf(existing).Kind()
			if existingKind != reflect.Slice && existingKind != reflect.Array {
				slice = []any{existing}
			} else {
				// Try to assert directly to []any.
				if s, ok := existing.([]any); ok {
					slice = s
				} else {
					v := reflect.ValueOf(existing)
					slice = make([]any, v.Len())
					for i := 0; i < v.Len(); i++ {
						slice[i] = v.Index(i).Interface()
					}
				}
			}

			// Check if the new value is an array/slice.
			if a.value != nil {

				newValType := reflect.TypeOf(a.value)
				if newValType.Kind() == reflect.Slice || newValType.Kind() == reflect.Array {
					v := reflect.ValueOf(a.value)
					for i := 0; i < v.Len(); i++ {
						slice = append(slice, v.Index(i).Interface())
					}
				} else {
					slice = append(slice, a.value)
				}
				list[a.key] = slice
			}
		} else {
			// If key doesn't exist, check if the value is an array/slice and convert it.

			if a.value != nil {

				valType := reflect.TypeOf(a.value)
				if valType.Kind() == reflect.Slice || valType.Kind() == reflect.Array {
					v := reflect.ValueOf(a.value)
					arr := make([]any, v.Len())
					for i := 0; i < v.Len(); i++ {
						arr[i] = v.Index(i).Interface()
					}
					list[a.key] = arr
				} else {
					list[a.key] = a.value
				}

			} else {
				list[a.key] = ""
			}
		}
	}

	return list

}

func New(in ...any) (string, error) {

	result, err := json.Marshal(Convert(in))
	if err != nil {
		return "", err
	}

	return string(result), nil
}
