// Package jts builds compact JSON objects for htmx hx-vals attributes.
//
// [New] is the primary API. It intentionally follows the alternating key/value
// style used by log/slog. Unlike slog.Info and similar logging functions, New has
// no leading message argument: its first string is a key. It returns ordinary
// JSON for use with templ's normal dynamic-attribute escaping:
//
//	hx-vals={ jts.New("key", value) }
//
// The result is data, not trusted HTML or JavaScript. Do not concatenate it into
// raw HTML, bypass templ escaping, or prefix it with htmx's js: or javascript:.
// Values submitted by a browser remain client-controlled and must be validated
// and authorized by the server.
//
// [Pair] is the optional explicit-pair form, analogous to using slog.Attr. The
// permissive [New] API remains canonical; the package does not provide a separate
// typed entry point.
//
// # Input forms
//
// Inputs are processed positionally from left to right. The supported forms are:
//
//   - A string key followed by any value. The string always consumes the next
//     argument as its value, even when that argument is another supported form.
//   - An [Input] constructed by [Pair].
//   - Exactly map[string]any, map[string]string, map[string][]any, or
//     map[string][]string. Each map entry becomes an input pair.
//   - A [url.Values] map. Each entry becomes an input pair.
//   - Exactly []any, which is processed as an argument list when it appears as
//     a standalone argument. A []any following a string key is an ordinary JSON
//     array value.
//
// Other standalone values, including defined string and map types, are stored
// under [BadKey]. A trailing string with no value is also stored under BadKey.
// url.Values is supported explicitly despite being a defined map type.
//
// Standalone []any argument lists are expanded in place, so arguments following
// an empty, nil, or populated argument list are still processed. Parsing is
// iterative. A cycle in standalone argument lists is treated as an unsupported
// standalone value under [BadKey]; [New] then reports the encoding/json cycle
// error instead of recursing without bound.
//
// # Compatibility behavior
//
// Empty keys and the zero value of [Input] are consumed but omitted. A lone
// empty string is instead an unmatched value and is stored as BadKey with an
// empty-string value. BadKey is an ordinary key, so explicit and generated
// BadKey entries aggregate like any other duplicate key.
//
// A value supplied once retains its encoding/json representation, including
// nested maps, structs, arrays, slices, pointers, json.RawMessage, custom
// marshalers, and []byte's base64 encoding. Nil values use encoding/json
// semantics and encode as null; empty non-nil slices encode as empty arrays.
//
// Duplicate values are combined in encounter order. Non-nil slices and arrays
// concatenate by one level so repeated homogeneous collections can unmarshal
// into a single Go slice. Nested slices remain nested. Byte slices and values
// implementing json.Marshaler or encoding.TextMarshaler are treated atomically
// to preserve their JSON representation. Supplied maps and structs are also
// atomic and are never deep-merged.
//
// [Convert] always returns a non-nil map and does not verify that its values can
// be encoded as JSON. [New] returns an empty string and the encoding/json error
// for unsupported values, non-finite floats, cycles, or custom marshaling errors.
// Convert map iteration order is unspecified. New returns compact JSON whose
// object-key ordering and escaping are determined by encoding/json. The package
// guarantees the attribute JSON, not the request representation that htmx and a
// browser later choose for those values.
package jts
