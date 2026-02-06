# Agent-Friendly Improvements Plan

Scope: implement the remaining agent-friendly upgrades discussed:
1. Machine-readable help/manifest
2. Retry/backoff + timeouts (configurable)
3. `--jsonl` streaming output
4. More “flat access” resolvers for nested resources
5. Normalized success envelope in agent mode
6. Structured agent-mode errors for local validation (not just API errors)
7. Agent-safe behavior for interactive prompts

## 1) Machine-Readable Help/Manifest
- Add `deel meta` command group:
  - `deel meta commands` outputs the full command tree as JSON.
  - `deel meta help [command path...]` outputs a single command’s schema (use, flags, subcommands) as JSON.
- Make `deel --agent --help` emit JSON (same schema as `meta help`).
- Ensure non-agent help remains unchanged.

## 2) Configurable Retries + Timeouts
- Expose flags:
  - `--timeout` (default 30s)
  - `--retries` (default 3)
  - `--retry-base` (default 1s)
  - `--retry-max` (default 30s)
- Wire flags into `internal/api.Client` retry/backoff logic for both JSON and multipart requests.
- Keep retry behavior conservative for client errors (only retry 429/5xx + network).

## 3) JSONL Streaming Output
- Add persistent flag `--jsonl`:
  - Requires JSON output (auto-enables JSON if needed).
  - In JSONL mode, output one JSON object per line (compact).
  - Apply `--query/--jq` per item when possible.
- Implement centrally in `outfmt.Formatter.OutputFiltered` so it applies to all list outputs.

## 4) More “Flat Access” Resolvers
- Add `deel tasks get <task-id> [--contract-id <id>]`:
  - If contract id missing, resolve via `FindTaskContract`.
  - Fetch task via API if possible; otherwise list within contract and locate.

## 5) Normalized Agent Success Envelope
- In `--agent` mode (and only when NOT using `--query/--raw/--items/--jsonl`):
  - Wrap stdout JSON as `{ "ok": true, "result": <existing output> }`.
  - Keep agent errors as `{ "ok": false, "error": ... }`.

## 6) Structured Agent Errors For Local Validation
- Add `failValidation(...)` helper for required flags / invalid inputs:
  - Human-readable error + suggestions to stderr.
  - Structured JSON error on stdout in `--agent` mode (exactly one per process).
- Ensure wrapped errors still categorize correctly:
  - Use `errors.As` for `StatusCoder` and `Messager` so `%w` wrapping preserves category/message.
- For “flat access” lookups (like tasks):
  - Return `api.APIError{StatusCode:404,...}` for local not-found so agent errors classify as `not_found`.

## 7) Agent-Safe Prompts
- In `--agent` mode:
  - Disallow interactive confirmations unless `--force` is provided.
  - Emit structured JSON validation error explaining how to proceed.

## Validation
- `go test ./...`
- Manual spot checks:
  - `deel --agent --help`
  - `deel --agent meta commands`
  - `deel --agent version`
  - `deel --agent timesheets delete-entry 123` (should be structured error, exit 1)
  - `deel --agent people list --jsonl | head`
