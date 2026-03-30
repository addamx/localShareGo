---
name: go
description: Pragmatic Go development skill for editing, debugging, and extending Go applications. Use when working on Go source files, module dependencies, cross-platform files, application state, HTTP handlers, or backend logic in repositories that use go.mod and targeted Go toolchain commands.
---

# Go

Use targeted Go workflows. Prefer reading `go.mod`, the nearest package files, and any repo instructions before editing.

For this repository:

- Treat `go.mod` as the source of truth for versions. The current baseline is `go 1.23.0`.
- Expect cross-platform files such as `*_windows.go` and `*_other.go`; preserve build-tag intent and keep platform-specific behavior split correctly.
- Prefer small, targeted verification such as `go test` on relevant packages or `go test ./...` only when it is justified.
- Do not introduce new tests, docs, or unrelated refactors unless the user asks.
- If frontend behavior is affected, coordinate with the `wails` skill instead of guessing the desktop bridge.

When changing code:

1. Read the package entry points and adjacent files before editing.
2. Preserve exported API shapes unless the user asked for a breaking change.
3. Keep error paths explicit and actionable.
4. Prefer standard library solutions before adding dependencies.
5. Update imports, structs, and call sites together so the package remains coherent.

Useful local checks:

- `go test ./...`
- `go test ./path/to/package`
- `go test -run <Name> ./path/to/package`
- `go fmt ./...` only when formatting is actually needed
