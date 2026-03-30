---
name: wails
description: Wails desktop application skill for repositories that use Wails v2 with a Go backend and a frontend folder. Use when changing bindings, startup or shutdown flow, asset embedding, frontend-backend interaction, wails.json configuration, or desktop runtime behavior.
---

# Wails

Start from the actual project wiring instead of generic Electron-style assumptions.

For this repository:

- `main.go` runs `wails.Run(...)`, binds `app`, and embeds `frontend/dist`.
- `wails.json` defines the frontend commands; do not proactively run `npm run build`.
- Frontend sources live under `frontend/`; backend bindings and lifecycle logic live in Go files at the repo root.
- The current Wails dependency is `github.com/wailsapp/wails/v2 v2.12.0`.

When working on Wails tasks:

1. Inspect `main.go`, `wails.json`, and the bound app methods before editing.
2. Keep the contract between Go-exposed methods and frontend callers synchronized.
3. Prefer fixing the actual startup, runtime, or asset flow rather than adding workaround state.
4. If a change touches LAN sharing or HTTP serving, verify the real listener state and self-check paths instead of assuming token or UI issues first.
5. When frontend behavior depends on backend state, trace the whole path from Wails binding to UI consumption.

Repo-specific cautions:

- Do not assume `http server status=running` means the service is reachable; startup failures and thread panics must be surfaced explicitly.
- On multi-NIC machines, prefer exposing multiple candidate LAN addresses with self-check results.
- If LAN HTTP fails, inspect in-app diagnostics and actual listening state before blaming generated links.

Useful local checks:

- Read `wails.json` before invoking any Wails command.
- Use targeted Go checks for backend changes.
- Use frontend package metadata to understand available scripts instead of assuming a framework default.
