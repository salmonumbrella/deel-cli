# Deel CLI Refactor Plan (2026-01-10)

Goal: reduce repetition, make pagination/output consistent, and centralize API decoding so commands are simpler and safer to extend.

## 1) Standardize command plumbing + error handling
- Add shared helpers in `internal/cmd` to centralize:
  - client + formatter creation
  - dry-run handling for mutations
  - consistent `HandleError` usage
- Replace ad‑hoc `PrintError` + `return err` branches with `HandleError` so suggestions and formatting are consistent.
- Fix stale suggestions (e.g., references to commands that don’t exist).

## 2) Unify pagination and list output
- Replace repeated cursor loops with a shared cursor pagination helper.
- Centralize the “More results available…” message.
- Ensure list output goes through the same output path (text + JSON) with consistent filtering.

## 3) Make API decoding generic and DRY
- Introduce generic response helpers in `internal/api`:
  - `DataResponse[T]`, `ListResponse[T]`, `Page`
  - `decodeData`, `decodeList`, and `wrapData`
- Migrate API endpoints to use the helpers (reduce inline `json.Unmarshal` wrappers).

## 4) Align output formatting with global flags
- Move all command output to `Formatter.OutputFiltered(ctx, ...)` so `--json`, `--data-only`, and `--jq` behave consistently.
- Keep `getFormatter` defaults intact, but ensure context is honored.

## 5) Tests for new abstractions
- Update list helper tests for cursor pagination behavior.
- Add/adjust tests for response helpers where relevant.
- Ensure gofmt and existing tests still pass.

---

Implementation will follow the steps above in order to keep changes mechanical and reviewable.
