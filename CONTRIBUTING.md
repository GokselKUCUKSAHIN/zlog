# Contributing to zlog

Thanks for your interest in contributing to **zlog** — a lightweight, fluent wrapper around Go's `log/slog`.

By participating, you agree to follow our [Code of Conduct](CODE_OF_CONDUCT.md).

## Getting started

1. Fork the repository and clone your fork.
2. Install **Go 1.21 or newer**.
3. From the repository root, verify the suite passes:

```bash
go test ./...
```

Or use the Makefile:

```bash
make test
```

## How to contribute

1. **Open an issue** for bugs, API ideas, or larger changes before investing substantial work.
2. **Discuss** approach and scope when the change is non-trivial.
3. **Open a pull request** against the default branch with a clear description of *why* the change is needed.

Small, focused PRs are easier to review than large mixed patches.

## Development standards

### Formatting and style

- Format all Go code with `gofmt`.
- Prefer clear, idiomatic Go. Match the style of existing code in `zlog.go` and `internal/`.
- Comments and documentation for exported APIs should be in **English**.

### Zero dependencies

zlog is built on the Go standard library only (plus this repo's `internal/` package).

- Do **not** add third-party module dependencies.
- If an external library seems necessary, open an issue first and get maintainer agreement.

### Public API

- Keep the **fluent builder** style (`ZLogger` method chaining).
- Prefer existing naming conventions (`Err` / `WithError`, `Msg` / `Message`, `Msgf` / `Messagef`, `Segment`, `KeyValue`, etc.).
- Avoid breaking changes to the public API unless the PR clearly documents them and there is a strong reason.
- Put non-exported helpers in `internal/` when they are not part of the public surface.

## Testing

The project uses **black-box unit tests** (`package zlog_test`). Output is captured with `io.Writer` / `bytes.Buffer` for fast, isolated runs.

For full detail, see [TEST_README.md](TEST_README.md). Summary rules:

- Use `setupTestLogger` and a **fresh buffer per test** (no shared mutable output state).
- Prefer table-driven tests with `t.Run` where it improves clarity.
- Name regression tests with the `TestRegression` prefix.
- Name edge-case tests with the `TestEdgeCase` prefix.
- For new features: add a unit test, a regression test for interaction with existing behavior, and a benchmark when performance may change.
- Aim to keep statement coverage at **≥ 80%** (`make test-coverage`).
- `Fatal` / `Fatalf` call `os.Exit` and are intentionally not unit-tested in-process.

Useful commands:

```bash
make test                 # Run tests
make test-verbose         # Tests with coverage summary
make test-coverage        # Function-level coverage report
make test-coverage-html   # HTML coverage report
make bench                # Benchmarks
make test-race            # Race detector
make test-all             # Verbose + coverage + bench + race
```

There is no PR CI workflow yet; please run tests locally (ideally `make test-all`) before requesting review.

## Pull requests

Before opening a PR, please confirm:

- [ ] Code is formatted with `gofmt`
- [ ] `go test ./...` (or `make test-all`) passes
- [ ] Coverage did not meaningfully drop below the project target
- [ ] New behavior has tests (and regression / edge-case coverage when relevant)
- [ ] PR description explains the problem and the solution
- [ ] Breaking API changes are called out explicitly

## Conduct and security-sensitive reports

- Community standards: [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md)
- Report CoC violations privately via GitHub by contacting [@GokselKUCUKSAHIN](https://github.com/GokselKUCUKSAHIN)
- Security vulnerabilities: see [SECURITY.md](SECURITY.md) (private report only; do not file public issues)

## License

By contributing, you agree that your contributions will be licensed under the project's [MIT License](LICENSE).
