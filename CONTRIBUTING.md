# Contributing to JTS

Thanks for helping improve JTS. The project aims to remain a small, dependency-free,
and auditable Go package.

## Before starting

Read `AGENTS.md` for repository-wide engineering and security rules. For a bug fix,
first add a test that demonstrates the current behavior. Discuss changes to exported
APIs, accepted inputs, output semantics, dependencies, or the minimum Go version
before implementing them.

Security vulnerabilities must not be reported in a public issue. Follow
`SECURITY.md` instead.

## Development environment

JTS supports the oldest of the two Go major releases that receive upstream security
fixes. That is currently Go 1.26. Use the newest patch release available for your
chosen supported major.

Clone the repository and run the baseline checks:

```sh
go vet ./...
go test -count=1 ./...
go test -race -count=1 ./...
```

Format Go files before submitting a change:

```sh
gofmt -w *.go
```

Changes to input conversion, reflection, or JSON serialization should also exercise
the fuzz target:

```sh
go test -run=^$ -fuzz=FuzzNew -fuzztime=30s ./...
```

Run the representative conversion and serialization benchmarks before and after a
performance-sensitive change. Compare results on the same machine and Go version:

```sh
go test -run=^$ -bench=BenchmarkNew -benchmem ./...
```

Development tools and their transitive dependencies are pinned separately in
`tools/go.mod` and `tools/go.sum`, so they do not affect users of the JTS module.
Run them through the Go toolchain from the repository root; no global installation
is needed:

```sh
go tool -modfile=tools/go.mod govulncheck ./...
go tool -modfile=tools/go.mod staticcheck ./...
```

After an approved tool update, run:

```sh
go get -modfile=tools/go.mod -tool <package>@<version>
go mod tidy -modfile=tools/go.mod
```

Commit both tool module files with the update.

## Design expectations

- Preserve existing behavior unless a deliberate change has been approved.
- Prefer the standard library; justify every new dependency.
- Keep the package flat and APIs concrete.
- Add black-box, table-driven tests for behavior changes.
- Include adversarial cases for untrusted keys and values.
- Keep `hx-vals` output as non-executable JSON and use templ's normal attribute
  escaping in examples.
- Never treat values returned by a browser as authoritative. Validation,
  authentication, authorization, and CSRF protection belong at the server boundary.

## Pull requests

Keep each pull request focused. Its description should state:

- the problem and intended behavior;
- compatibility and security effects;
- tests and analysis commands run;
- any decision requiring maintainer approval;
- documentation updated or intentionally left unchanged.

JTS uses Conventional Commits because Release Please derives versions and GitHub
release notes from commits on `master`. Pull requests are normally squash-merged,
so give the pull request a release-ready title in the form `type(scope): summary`.
Use `fix` or `perf` for a v0 patch, `feat` for a v0 minor, and `!` plus a
`BREAKING CHANGE:` body for an approved breaking change. Documentation, tests,
refactoring, CI, and maintenance use `docs`, `test`, `refactor`, `ci`, and `chore`
respectively and do not release by themselves.

During v0, breaking changes target a new minor release and require explicit
approval plus migration guidance in the commit body; patch releases preserve
documented APIs and behavior. After v1, changes must preserve compatibility across
the v1 module line.

Before requesting review, confirm that formatting is clean, tests pass on the
minimum supported Go version, exported identifiers have useful doc comments, and no
credentials or generated fuzz artifacts are included.

## Releases

The Release Please workflow runs after pushes to `master`. Releasable commits create
or update a release pull request containing the next version and generated notes.
Merging that pull request creates the immutable `v`-prefixed tag and corresponding
GitHub Release. If no `fix`, `perf`, `feat`, or approved breaking commit has landed,
no release is proposed.

Repository administrators must allow GitHub Actions to create pull requests in the
repository's Actions settings. The workflow otherwise uses only the repository's
short-lived `GITHUB_TOKEN`; it does not require a personal access token.

Do not edit `.release-please-manifest.json` outside the initial bootstrap, create
release tags manually, or replace an existing tag. Retry a transient automation
failure with the workflow's manual dispatch instead.
