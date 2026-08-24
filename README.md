# JTS

[![Go Reference](https://pkg.go.dev/badge/github.com/hcarriz/jts.svg)](https://pkg.go.dev/github.com/hcarriz/jts)
[![Go Report Card](https://goreportcard.com/badge/github.com/hcarriz/jts)](https://goreportcard.com/report/github.com/hcarriz/jts)

JTS populates the [`hx-vals`](https://htmx.org/attributes/hx-vals/) attribute in
[templ](https://github.com/a-h/templ) projects. Inspired by
[`slog`](https://pkg.go.dev/log/slog), it accepts key/value strings, maps, and
values created by `jts.Pair`. Nested values follow Go's `encoding/json` rules.

The API intentionally mirrors the key/value portion of calls such as
`slog.Info("message", "key", value)`: `jts.New("key", value)` returns the
corresponding JSON object as a string. JTS has no leading message argument; its
first string is a key.

JTS requires Go 1.26 or newer. The minimum follows the oldest of the two Go major
releases that currently receive upstream security fixes.

## Installation

```sh
go get github.com/hcarriz/jts
```

## Usage

```templ
package button

import "github.com/hcarriz/jts"

templ Button() {
    <button hx-post="/my/specific/endpoint" hx-vals={ jts.New("key", "value") }>Submit</button>
}
```

`jts.New` is the primary API. Its current input forms are:

| Input | Current behavior |
| --- | --- |
| `"key", value` | A string consumes the following argument as its value. |
| `jts.Pair("key", value)` | Creates an opaque `jts.Input` pair. |
| Supported maps | Exactly `map[string]any`, `map[string]string`, `map[string][]any`, and `map[string][]string`. |
| `url.Values` | Adds each entry as a pair. |
| `[]any` | As a standalone argument, expands in place as an argument list. As a keyed value, remains a JSON array. |
| Other standalone values | Stores the value under `jts.BadKey`. |

`New(...any)` is intentionally the canonical API so JTS retains slog-like
key/value ergonomics. `Pair` is the optional explicit-pair form, analogous to
using `slog.Attr`; the project does not provide a separate typed entry point.

Values supplied once retain their normal `encoding/json` representation, including
nested maps and structs, `[]byte`, `json.RawMessage`, and custom marshalers.
Duplicate values retain encounter order; non-nil arrays and slices concatenate by
one level, while byte slices, custom-marshaled values, maps, and structs remain
atomic. Maps and structs are collected without deep-merging. Nil values encode as
JSON `null`. Empty keys are omitted.

Nested values can be decoded into matching Go types:

```go
type Filter struct {
	Status string   `json:"status"`
	Tags   []string `json:"tags"`
}

attribute, err := jts.New(
	"filter", Filter{Status: "open", Tags: []string{"go", "security"}},
	"ids", []int{1, 2},
	jts.Pair("ids", []int{3, 4}),
)
if err != nil {
	return err
}

var values struct {
	Filter Filter `json:"filter"`
	IDs    []int  `json:"ids"`
}
if err := json.Unmarshal([]byte(attribute), &values); err != nil {
	return err
}
```

This unmarshalling example describes JTS's JSON output. An `hx-vals` attribute is
not itself a JSON request body: htmx, the browser's request encoding, and the
server's form or body decoder determine the eventual wire representation.

### Pre-v1 compatibility note

Nested-value support intentionally changes earlier behavior: nil values now encode
as `null`; byte slices and custom marshalers keep their `encoding/json` form; and a
standalone `[]any` no longer discards later outer arguments. Duplicate collection
ordering and one-level array concatenation remain unchanged.

See [the package documentation](https://pkg.go.dev/github.com/hcarriz/jts) for the
authoritative behavior contract, compatibility details, and more examples.

## Safe use

`jts.New` produces ordinary JSON; it does not make request data trustworthy. Pass
its result through templ's normal dynamic-attribute syntax, as shown above. Do not
concatenate it into raw HTML, bypass templ escaping, or add htmx's `js:` or
`javascript:` prefix.

Treat every value received from `hx-vals` as client-controlled. Never place secrets
in the attribute or trust submitted values for identity, authorization, pricing, or
other security decisions. Validate and authorize them on the server and protect
state-changing endpoints against CSRF. Decode nested parameter values explicitly
according to the request content type; do not assume htmx submitted JTS's complete
object as a JSON body.

See [SECURITY.md](SECURITY.md) for the full threat model and private reporting
instructions.

## Versioning

Release Please creates the initial `v0.1.0` tag and subsequent releases from
Conventional Commits on `master`. During v0, `fix` and `perf` commits produce patch
releases; `feat` commits and approved breaking changes produce minor releases. A
breaking change must include migration guidance. Beginning with v1, JTS will follow
Go's compatibility promise for the entire v1 module line.

See [GitHub Releases](https://github.com/hcarriz/jts/releases) for release notes.

## Contributing

Repository conventions and required checks are documented in
[AGENTS.md](AGENTS.md). See [CONTRIBUTING.md](CONTRIBUTING.md) before submitting a
change.
