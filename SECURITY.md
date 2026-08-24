# Security policy

## Supported versions

Security fixes are made on the default branch and on the latest published release
when one exists. Older pre-v1 revisions are not guaranteed to receive backports.

JTS follows the Go project's supported-release window: the oldest supported Go
version is the older of the two major releases receiving upstream security fixes.
JTS currently requires Go 1.26. Users should run the latest patch release available
for their supported Go major.

## Reporting a vulnerability

Do not disclose a suspected vulnerability in a public issue, discussion, pull
request, test case, or commit.

Use [GitHub private vulnerability reporting][private-report] to send:

- the affected revision or version;
- a minimal reproduction or proof of concept;
- the expected impact and attack preconditions;
- suggested mitigations, if known;
- whether the report or related details have been shared elsewhere.

If private reporting is unavailable, contact the maintainer privately through the
contact information on the maintainer's GitHub profile. Do not include exploit
details in a public message. Reports will be investigated before coordinated
disclosure; no response or remediation deadline is promised.

## Security model

JTS serializes Go values to ordinary JSON for an HTML `hx-vals` attribute. It does
not authenticate, authorize, validate application data, sanitize HTML, implement
CSRF protection, or make client-provided values trustworthy.

JTS guarantees only the generated attribute JSON. It does not control the request
representation produced later by htmx and the browser. In particular, an
`hx-vals` attribute is not automatically a JSON request body. Server handlers must
decode the actual request content type and explicitly decode any nested parameter
values their application accepts.

The safe integration pattern is:

```templ
templ Button(userValue string) {
    <button hx-post="/endpoint" hx-vals={ jts.New("value", userValue) }>
        Submit
    </button>
}
```

This pattern relies on two independent encoders:

1. JTS uses `encoding/json` to create valid, non-executable JSON.
2. templ contextually HTML-escapes the resulting dynamic attribute value.

Do not concatenate the result into raw HTML, disable templ escaping, mark it as
trusted HTML, use an unquoted attribute, or prefix it with `js:` or `javascript:`.
htmx evaluates those prefixes as JavaScript, which can turn user-controlled data
into cross-site scripting.

## Application responsibilities

- Assume every `hx-vals` name and value can be read and changed in browser developer
  tools or by a forged HTTP request.
- Do not put secrets in `hx-vals`.
- Do not trust it for user identity, authorization, permissions, prices, ownership,
  redirect targets, or other sensitive decisions.
- Validate types, lengths, formats, and allowed values on the server. Enforce
  authentication and authorization independently of submitted values.
- Decode nested values according to the received request representation; do not
  assume Go's `json.Unmarshal` can be applied directly to an htmx form body.
- Protect state-changing endpoints against CSRF using an application-appropriate
  design, such as same-site cookies plus a validated CSRF token or origin checks.
- Bound the size and nesting of data before passing attacker-influenced structures
  to JTS when resource exhaustion is a concern.
- Prefer htmx's default same-origin request restriction. Applications that do not
  use JavaScript-evaluated htmx features should consider setting
  `htmx.config.allowEval` to `false` and deploying a restrictive Content Security
  Policy as defense in depth.

## Maintainer requirements

Security-sensitive changes require regression tests and fuzzing. Before release,
maintainers should run `go vet`, the race detector, and the pinned analysis tools:

```sh
go tool -modfile=tools/go.mod govulncheck ./...
go tool -modfile=tools/go.mod staticcheck ./...
```

Dependency additions and changes to escaping, input semantics, or generated output
require explicit review.

The required CI workflow runs the vulnerability scan and a short fuzz pass. Before
a release or after a security-sensitive change, manually dispatch the `Maintenance`
workflow for its ten-minute fuzz pass and repeat vulnerability scan.

## References

- [Go release and security-fix policy][go-releases]
- [templ dynamic attribute escaping][templ-attributes]
- [htmx `hx-vals` security considerations][htmx-vals]
- [htmx web security guidance][htmx-security]

[private-report]: https://github.com/hcarriz/jts/security/advisories/new
[go-releases]: https://go.dev/doc/devel/release
[templ-attributes]: https://templ.guide/syntax-and-usage/attributes/
[htmx-vals]: https://htmx.org/attributes/hx-vals/
[htmx-security]: https://htmx.org/docs/#web-security
