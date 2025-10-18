# Repository Guidelines

## Project Structure & Module Organization
The CLI entry point lives in `cmd/ip-fetcher`, with one command file per provider plus `main.go`. Reusable provider clients live in `providers/<name>`, each exposing a `Fetch()` implementation shared by the CLI and publisher workflow. Shared helpers (logging, HTTP, formatting) sit under `internal/` (`internal/web`, `internal/output`, `internal/pflog`), while `fetchers/` holds the generic web fetcher powering multiple providers. Release automation and git publishing helpers reside in `publisher/`, and generated binaries land in `dist/`; the `dist/` directory is ignored by git and should remain out of commits.

## Build, Test, and Development Commands
Use `make build` for a local CLI binary in `dist/ip-fetcher`; `make build-all` produces cross-platform artifacts and injects version metadata. `make test` runs `go test` with coverage aggregation into `coverage.txt`. `make lint` executes `golangci-lint`, `make fmt` rewrites sources via `gofumpt`, and `make ci` chains lint plus tests for PR validation. During rapid iteration, `go test ./cmd/ip-fetcher -run TestAWS` targets a single provider suite, while `make critic` runs `gocritic` checks when deeper static analysis is needed.

## Coding Style & Naming Conventions
Target Go 1.24 and the standard Go formatting conventions; indent with tabs and keep statements idiomatic. Run `make fmt` before committing and let `golangci-lint` cover imports, vetting, and static analysis. Provider files and packages mirror the upstream service (e.g., `providers/aws`, `cmd/ip-fetcher/aws.go`); keep mocks or fixtures alongside tests in matching `_test.go` files.

## Testing Guidelines
Write table-driven tests in the same package using `_test.go` suffixes; see `cmd/ip-fetcher/aws_test.go` as a pattern. Prefer `stretchr/testify` assertions and `gock` for HTTP fixtures, and reset shared state between cases. The generated `coverage.txt` is ignored by git but should stay deterministic so CI comparisons remain meaningful. When adding a provider, include integration-safe tests that stub remote calls instead of hitting the network.

## Commit & Pull Request Guidelines
Follow the existing history with short, descriptive commit summaries (e.g., `reduce repetition of code.` or `Bump github.com/stretchr/testify to 1.11.1`) and add body text when context helps; keep lines under ~72 characters. Reference issues with `Fixes #123` when applicable. Pull requests should describe the change, outline manual or automated tests run (linking to `make ci` output when relevant), and note any configuration or documentation updates. Attach screenshots only when altering user-facing output (e.g., new CLI flags). Update `README.md` or provider docs whenever behavior or flags change.
