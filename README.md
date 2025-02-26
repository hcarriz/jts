# JTS

[![Go Reference](https://pkg.go.dev/badge/github.com/hcarriz/jts.svg)](https://pkg.go.dev/github.com/hcarriz/jts)
[![Go Report Card](https://goreportcard.com/badge/github.com/hcarriz/jts)](https://goreportcard.com/report/github.com/hcarriz/jts)

The goal of the project is to populate the hx-vals attribute in [templ](https://github.com/a-h/templ) projects. Inspired by [slog](https://pkg.go.dev/log/slog), you can pass strings, maps, and `jts.Input`.

> ![WARNING]
>
> It is not recommended to use nested entities.

## Installation

```
go get github.com/hcarriz/jts
```

## Usage

```templ
package button

import "github.com/hcarriz/jts"

templ Button() {
    <button hx-post="/my/specific/endpoint" hx-vals={ jts.New("key","value") }>Submit</button>
}
```

Please see [godoc](https://pkg.go.dev/github.com/hcarriz/jts) for more examples on how to use `jts.New()`.
