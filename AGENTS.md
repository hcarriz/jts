# Repository guidance for coding agents

## Scope

These instructions apply to the entire repository. JTS is a small, dependency-free
Go library that converts a permissive list of values into JSON for htmx `hx-vals`
attributes, primarily in templ components.

Keep the package small, unsurprising, and easy to audit. Prefer the Go standard
library and direct code over frameworks, generated abstractions, or additional
packages.

## Change authority

Preserve backward compatibility where practical. The module is pre-v1, so a
breaking change can be considered, but it is never implicit.

Get explicit maintainer approval before changing any of the following:

- an exported identifier or its signature;
- accepted inputs, output shape, duplicate-key ordering, nil handling, error
  behavior, or another observable behavior;
- the minimum Go version;
- a runtime or development dependency;
- the security model described in `SECURITY.md`;
- release, licensing, or repository automation policy.

General modernization approval is not approval to make one of these changes.
Present the specific decision and its compatibility and security effects first.

## API and release policy

`New(...any)` is intentionally the canonical API. It mirrors the alternating
key/value portion of `log/slog` calls while returning a JSON string; unlike
`slog.Info`, it has no leading message argument. `Pair` is the optional
explicit-pair form. Do not add a parallel typed entry point without a new,
maintainer-approved use case.

Release Please creates release pull requests, `v`-prefixed tags, and GitHub Releases
from Conventional Commits on `master`. The initial version is `v0.1.0`. During v0,
`fix` and `perf` commits produce patch releases, while `feat` commits and approved
breaking changes produce minor releases. Include `BREAKING CHANGE:` migration
guidance in the commit body for every approved break. Commits such as `docs`,
`test`, `refactor`, `ci`, and `chore` do not release by themselves.

Beginning with v1, preserve source compatibility throughout the v1 module line in
accordance with Go module versioning. Never delete or retag a published version;
roll fixes forward and use a `retract` directive when withdrawal is necessary.
GitHub Releases are the release-note source of truth; do not add a generated
changelog or commit release-time version changes outside Release Please's manifest.

## Go version policy

Support the oldest of the two Go major releases that currently receive upstream
security fixes. The module currently declares Go 1.26. When that support window
changes, propose the version bump separately and update `go.mod`, CI, `README.md`,
and contributor documentation together.

Use the latest patch release of each supported Go major in development and CI.
Use language and standard-library features available in the minimum version only.

## Security invariants

- `New` must produce ordinary valid JSON. Do not add a `js:` or `javascript:`
  prefix, executable expressions, or APIs that mark output as trusted HTML.
- Treat keys and values as untrusted. Preserve `encoding/json` escaping and require
  callers to place the returned string through templ's normal dynamic-attribute
  escaping, for example `hx-vals={ jts.New(...) }`.
- Never recommend string concatenation into HTML, unquoted attributes,
  `template.HTML`, `templ.Raw`, or another escaping bypass for `hx-vals`.
- Document that everything in `hx-vals` is visible to and modifiable by the client.
  It must not contain secrets or be trusted for authorization, identity, pricing,
  access control, or other security decisions. Servers must validate and authorize
  every submitted value.
- Keep resource-exhaustion risks in mind for deeply nested, cyclic, or very large
  values. Add regression or fuzz coverage for any parser or reflection change.
- Do not log or commit credentials, tokens, private vulnerability reports, or
  generated artifacts that may contain them.

See `SECURITY.md` for the supported threat model and reporting process.

## Implementation style

- Write idiomatic, formatted Go with direct control flow and early error returns.
- Keep the single-package layout unless a new package has a clear independent
  responsibility.
- Prefer concrete types. Define small interfaces at their point of use only when
  multiple implementations are actually needed.
- Wrap errors with useful operation context while preserving the cause with `%w`.
- Avoid package-level mutable state. Every goroutine must have an explicit lifetime
  and cancellation path.
- Prefer standard-library functionality to third-party dependencies.
- Add doc comments to exported API. Comments must describe behavior and important
  edge cases rather than repeat the identifier name.
- Do not perform unrelated cleanup in a focused change.

## Development tools

Development commands are declared with Go `tool` directives and pinned in
`tools/go.mod` and `tools/go.sum`. The separate development module keeps tool
dependencies out of the library's public `go.mod` while using the conventional
module filenames understood by dependency automation.

Run tools from the repository root with `-modfile=tools/go.mod`. Do not add a
legacy `tools.go` file, rely on globally installed binaries, or use
`go run ...@latest` in documented or automated workflows.

Adding or updating a tool requires explicit approval. Make an approved update with:

```sh
go get -modfile=tools/go.mod -tool <package>@<version>
go mod tidy -modfile=tools/go.mod
```

Review and commit both tool module files after the checks pass.

## Testing

Add or update tests with every behavior change. Prefer table-driven tests and keep
black-box API tests in package `jts_test`. Cover malformed and adversarial inputs,
typed nils, arrays and slices, duplicate keys, unsupported JSON values, cycles, and
large inputs when relevant.

Run the checks appropriate to the change:

```sh
gofmt -w *.go
go vet ./...
go test -count=1 ./...
go test -race -count=1 ./...
```

For changes to conversion, reflection, or serialization, also run the fuzz target:

```sh
go test -run=^$ -fuzz=FuzzNew -fuzztime=30s ./...
```

For performance-sensitive changes, compare the representative benchmarks before
and after the change on the same machine and Go version:

```sh
go test -run=^$ -bench=BenchmarkNew -benchmem ./...
```

Run the pinned security and static-analysis tools:

```sh
go tool -modfile=tools/go.mod govulncheck ./...
go tool -modfile=tools/go.mod staticcheck ./...
```

Do not update a failing test merely to bless changed behavior unless that behavior
has been explicitly approved.

## Repository automation

CI resolves `oldstable` and `stable` through `actions/setup-go`; the minimum lane
must also verify that `go.mod` declares the resolved `oldstable` major.minor. Keep
the required job names `Test (minimum)`, `Test (current)`, `Quality`, and
`Fuzz smoke` stable because repository rules may refer to them.

Pin third-party actions to full commit SHAs. CI and maintenance workflows grant the
workflow token read-only repository contents and disable persisted checkout
credentials. Release Please is the narrow exception: its job may grant only
`contents: write` and `pull-requests: write` so it can open release pull requests,
create immutable tags, and publish GitHub Releases. Do not introduce
`pull_request_target` or custom secrets into pull-request checks. The extended fuzz
and vulnerability workflow is deliberately manual because this repository is
expected to have low activity; do not add a schedule without explicit approval.

## Documentation and handoff

- Keep examples executable and check returned errors in production examples.
- Update `README.md`, `SECURITY.md`, and `CONTRIBUTING.md` when their guarantees or
  workflows change.
- Record deferred hardening work in the issue tracker; do not leave vague TODO
  comments in production code.
- In the final handoff, summarize observable changes, approval-sensitive decisions,
  checks run, and any remaining security or compatibility concerns.
