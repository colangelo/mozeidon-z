<!-- OPENSPEC:START -->
# OpenSpec Instructions

These instructions are for AI assistants working in this project.

Always open `@/openspec/AGENTS.md` when the request:
- Mentions planning or proposals (words like proposal, spec, change, plan)
- Introduces new capabilities, breaking changes, architecture shifts, or big performance/security work
- Sounds ambiguous and you need the authoritative spec before coding

Use `@/openspec/AGENTS.md` to learn:
- How to create and apply change proposals
- Spec format and conventions
- Project structure and guidelines

Keep this managed block so 'openspec update' can refresh the instructions.

<!-- OPENSPEC:END -->

# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Mozeidon is a CLI tool that controls Firefox/Chrome browsers from the terminal via IPC and native messaging protocols. It manages tabs, bookmarks, history, and tab groups.

## Architecture

```
CLI (Go)  →  Native App (IPC)  →  Browser Extension (Native Messaging)  →  Browser APIs
```

Three components must work together:
1. **CLI** (`/cli`) - Go binary using Cobra, communicates via IPC to `mozeidon_native_app`
2. **Browser Extensions** (`/firefox-addon`, `/chrome-addon`) - TypeScript WebExtensions that receive commands and call browser APIs
3. **Native App** (separate repo: `mozeidon-native-app`) - IPC broker between CLI and extension

The CLI sends `Command` structs (command name + args string) via IPC. Extensions dispatch commands through `handler.ts` to service files (`tabs.ts`, `bookmarks.ts`, etc.) which call WebExtension APIs. Responses flow back through the same path.

## Beads workflow

Check ./BD_WORKFLOW.md

## MCP Agent Mail: coordination for multi-agent workflows

What it is
- A mail-like layer that lets coding agents coordinate asynchronously via MCP tools and resources.
- Provides identities, inbox/outbox, searchable threads, and advisory file reservations, with human-auditable artifacts in Git.

Why it's useful
- Prevents agents from stepping on each other with explicit file reservations (leases) for files/globs.
- Keeps communication out of your token budget by storing messages in a per-project archive.
- Offers quick reads (`resource://inbox/...`, `resource://thread/...`) and macros that bundle common flows.

How to use effectively
1) Same repository
   - Register an identity: call `ensure_project`, then `register_agent` using this repo's absolute path as `project_key`.
   - Reserve files before you edit: `file_reservation_paths(project_key, agent_name, ["src/**"], ttl_seconds=3600, exclusive=true)` to signal intent and avoid conflict.
   - Communicate with threads: use `send_message(..., thread_id="FEAT-123")`; check inbox with `fetch_inbox` and acknowledge with `acknowledge_message`.
   - Read fast: `resource://inbox/{Agent}?project=<abs-path>&limit=20` or `resource://thread/{id}?project=<abs-path>&include_bodies=true`.
   - Tip: set `AGENT_NAME` in your environment so the pre-commit guard can block commits that conflict with others' active exclusive file reservations.

2) Across different repos in one project (e.g., Next.js frontend + FastAPI backend)
   - Option A (single project bus): register both sides under the same `project_key` (shared key/path). Keep reservation patterns specific (e.g., `frontend/**` vs `backend/**`).
   - Option B (separate projects): each repo has its own `project_key`; use `macro_contact_handshake` or `request_contact`/`respond_contact` to link agents, then message directly. Keep a shared `thread_id` (e.g., ticket key) across repos for clean summaries/audits.

Macros vs granular tools
- Prefer macros when you want speed or are on a smaller model: `macro_start_session`, `macro_prepare_thread`, `macro_file_reservation_cycle`, `macro_contact_handshake`.
- Use granular tools when you need control: `register_agent`, `file_reservation_paths`, `send_message`, `fetch_inbox`, `acknowledge_message`.

Common pitfalls
- "from_agent not registered": always `register_agent` in the correct `project_key` first.
- "FILE_RESERVATION_CONFLICT": adjust patterns, wait for expiry, or use a non-exclusive reservation when appropriate.
- Auth errors: if JWT+JWKS is enabled, include a bearer token with a `kid` that matches server JWKS; static bearer is used only when JWT is disabled.

## Build Commands

```bash
# Build everything
make all

# Build individual components
make build-cli              # Builds Go binary at cli/mozeidon
make build-firefox-addon    # Runs npm install, prettier, webpack in firefox-addon/
make build-chrome-addon     # Runs npm install, prettier, webpack in chrome-addon/

# Run CLI (after build)
./cli/mozeidon tabs get

# Extension development
cd firefox-addon && npm run prettier   # Format TypeScript
cd firefox-addon && npm run build      # Webpack build only

# Raycast extension
cd raycast && npm run dev    # Development mode
cd raycast && npm run lint   # Lint check
```

## Key File Locations

**CLI (Go 1.21, Cobra)**
- `cli/cmd/` - Cobra command definitions (tabs, bookmarks, bookmark, history, groups)
- `cli/core/` - Business logic for each operation
- `cli/browser/core/browser-service.go` - IPC client wrapper
- `cli/browser/infra/ipc-client.go` - golang-ipc implementation

**Extensions (TypeScript, Webpack)**
- `*/src/app.ts` - Entry point, native messaging listener
- `*/src/handler.ts` - Command dispatcher (switch on CommandName enum)
- `*/src/services/` - Browser API wrappers (tabs.ts, bookmarks.ts, history.ts, groups.ts)
- `*/src/models/command.ts` - CommandName enum defining all 15 commands
- Firefox uses Manifest V2 (`background.scripts`), Chrome uses Manifest V3 (service worker)

## Command Protocol

Commands are defined in `CommandName` enum:
- Tabs: `get-tabs`, `switch-tab`, `close-tabs`, `new-tab`, `update-tab`, `duplicate-tab`, `new-group-tab`, `get-recently-closed-tabs`
- Bookmarks: `get-bookmarks`, `write-bookmark`
- History: `get-history-items`, `delete-history-items`
- Groups: `get-groups`, `update-group`, `move-group`

Tab IDs use `windowId:tabId` format. Bookmark folder paths start and end with `/`.

## Testing Locally

1. Build: `make all`
2. Disable any installed mozeidon extension in browser
3. Load temporary extension: Firefox `about:debugging` → Load Temporary Add-on → select `firefox-addon/manifest.json`
4. Test: `./cli/mozeidon tabs get`

## Releases

Releases are automated via GitHub Actions + goreleaser. Push a tag to trigger:
```bash
git tag -a v2.0.0 -m "Release message"
git push origin v2.0.0
```
