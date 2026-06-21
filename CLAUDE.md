
# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Mozeidon-Z is a macOS-focused CLI tool that controls Firefox/Chrome browsers from the terminal via IPC and native messaging protocols. It manages tabs, bookmarks, history, and tab groups. It is a standalone hard fork of [`egovelox/mozeidon`](https://github.com/egovelox/mozeidon) (see *Keeping it current* below).

## Architecture

```
CLI (Go)  →  Native App (IPC)  →  Browser Extension (Native Messaging)  →  Browser APIs
```

Three components must work together:
1. **CLI** (`/cli`) - Go binary using Cobra, communicates via IPC to `mozeidon_native_app`
2. **Browser Extensions** (`/firefox-addon`, `/chrome-addon`) - TypeScript WebExtensions that receive commands and call browser APIs
3. **Native App** (separate repo: `mozeidon-native-app`) - IPC broker between CLI and extension

The CLI sends `Command` structs (command name + args string) via IPC. Extensions dispatch commands through `handler.ts` to service files (`tabs.ts`, `bookmarks.ts`, etc.) which call WebExtension APIs. Responses flow back through the same path.

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
just build-all

# Build individual components
just build-cli              # Builds Go binary at cli/mozeidon
just build-firefox          # npm install, prettier, webpack in firefox-addon/
just build-chrome           # syncs src from firefox-addon, then npm install + webpack

# Run CLI (after build)
./cli/mozeidon tabs get

# Extension development
cd firefox-addon && npm run prettier   # Format TypeScript
cd firefox-addon && npm run build      # Webpack build only

# Raycast extension
cd raycast && npm run dev    # Development mode
cd raycast && npm run lint   # Lint check
```

## Just Recipes

The project uses `just` (justfile) for common development tasks. Run `just --list` to see all available commands.

```bash
# Build
just build-all              # Build CLI + Firefox + Chrome addons
just build-cli              # Build CLI only
just build-firefox          # Build Firefox addon only
just build-raycast          # Build Raycast extension

# Setup (required for first-time setup)
just setup-native-messaging # Install Firefox native messaging manifest
just check-native-messaging # Verify native messaging is configured
just setup-all              # Full setup: build everything + configure native messaging

# CLI testing
just tabs-get               # Get open tabs
just tabs-closed            # Get recently closed tabs
just tabs-activate ID       # Activate a tab (e.g., just tabs-activate 3289:596)
just test-connection        # Test CLI can connect to Firefox

# Extension packaging (for AMO/Chrome Web Store upload)
just package-firefox        # Create mozeidon-firefox.xpi + mozeidon-source.zip
just package-chrome         # Create mozeidon-chrome.zip

# Development
just raycast-dev            # Run Raycast in dev mode
just format-firefox         # Format Firefox addon TypeScript
just firefox-debug          # Open Firefox debugging page (about:debugging)
```

## Key File Locations

**CLI (Go 1.24, Cobra)**
- `cli/cmd/` - Cobra command definitions (tabs, bookmarks, bookmark, history, groups)
- `cli/core/` - Business logic for each operation
- `cli/browser/core/browser-service.go` - IPC client wrapper
- `cli/browser/infra/ipc-client.go` - golang-ipc implementation

**Extensions (TypeScript, Webpack)**
- `*/src/app.ts` - Entry point, native messaging listener
- `*/src/handler.ts` - Command dispatcher (switch on CommandName enum)
- `*/src/services/` - Browser API wrappers (tabs.ts, bookmarks.ts, history.ts, groups.ts)
- `*/src/models/command.ts` - CommandName enum defining the command names
- Firefox uses Manifest V2 (`background.scripts`), Chrome uses Manifest V3 (service worker)

## Command Protocol

Commands are defined in `CommandName` enum:
- Tabs: `get-tabs`, `switch-tab`, `activate-tab`, `close-tabs`, `new-tab`, `update-tab`, `duplicate-tab`, `new-group-tab`, `get-recently-closed-tabs`
- Bookmarks: `get-bookmarks`, `write-bookmark`
- History: `get-history-items`, `delete-history-items`
- Groups: `get-groups`, `update-group`, `move-group`

Tab IDs use `windowId:tabId` format. Bookmark folder paths start and end with `/`.

## Testing Locally

1. Build: `just build-all`
2. Disable any installed mozeidon extension in browser
3. Load temporary extension: Firefox `about:debugging` → Load Temporary Add-on → select `firefox-addon/manifest.json`
4. Test: `./cli/mozeidon tabs get`

## CI / Releases (Woodpecker)

CI runs on **Woodpecker** (`ci.cat-bluegill.ts.net`), driven by `.woodpecker.yml`. The Gitea
repo `A-Layer/mozeidon-z` is activated there; pushes/tags fire pipelines.

- **Build pipeline** (`.woodpecker.yml`): mirrors `just build-all` — builds the CLI (`golang:1.24`)
  and both add-ons (`node:20`, chrome src synced from firefox-addon) in parallel on
  push/PR/manual. Raycast is excluded (its `ray build` needs the Raycast toolchain/login).
- **Release-on-tag → Gitea Releases:** planned, not yet wired. Needs a `gitea_token` Woodpecker
  secret — creating a Gitea Release is an API write the Woodpecker↔Gitea OAuth link alone
  doesn't cover.
- The old GitHub-Actions goreleaser path (`.github/workflows/release.yml`) is **deprecated**;
  disable **Gitea Actions** for the repo so it doesn't also pick up that workflow.
- Firefox extension distribution is unchanged: AMO via `just submit-firefox` (see below).

### Checking CI status (woodpecker-cli)

The agent can inspect and debug pipelines itself. Auth comes from 1Password — the token is in
the **AC-DevOps** vault (*not* `Private`, which is what some other repos' `.envrc` reference):

```bash
export WOODPECKER_SERVER=https://ci.cat-bluegill.ts.net
export WOODPECKER_TOKEN=$(op read 'op://AC-DevOps/woodpecker - Personal Access Token/password')

woodpecker-cli info                                            # verify auth (→ User: ac)
woodpecker-cli pipeline ls   A-Layer/mozeidon-z               # recent runs
woodpecker-cli pipeline last A-Layer/mozeidon-z               # latest run
woodpecker-cli pipeline ps   A-Layer/mozeidon-z <N>           # step states for run N
woodpecker-cli pipeline log show A-Layer/mozeidon-z <N> <step> # tail a step's log (debug)
woodpecker-cli pipeline start  A-Layer/mozeidon-z <N>          # re-run run N (same commit)
woodpecker-cli pipeline create A-Layer/mozeidon-z --branch main # fresh manual run
```

> **1Password / AFK caveat:** `op read` triggers a Touch ID / biometric prompt. If the user is
> away the prompt **times out** — that is **not** a failure and **not** a reason to abandon the
> task. Do not loop-retry; **wait for the user to return**, then re-run the command. (Probe op
> readiness with `op vault list`, never `op whoami`.)

## Keeping it current

**Mozeidon-Z** (`colangelo/mozeidon-z`) — a standalone hard fork of
[`egovelox/mozeidon`](https://github.com/egovelox/mozeidon) (`upstream` remote, kept only as an
occasional cherry-pick source). The CLI, both extensions, and the Raycast extension are built
from here; the native app comes from the separate `mozeidon-native-app` repo (or Homebrew).

**After changing the source, rebuild the affected piece — and note the loaded browser extension
does NOT auto-update, you must reload it:**

| Changed | Rebuild | Then |
|---|---|---|
| `cli/**` | `just install-cli` (or `just build-cli`) | re-run the CLI; nothing to reload |
| `firefox-addon/**` | `just build-firefox` | **dev:** `about:debugging` → Load Temporary Add-on. **release:** bump `manifest.json` version → `just package-firefox` → submit to AMO ([Mozeidon-Z](https://addons.mozilla.org/firefox/addon/mozeidon-z/)) → installs auto-update (release Firefox won't load an unsigned `.xpi`) |
| `chrome-addon/**` | `just build-chrome && just package-chrome` | reload at `chrome://extensions` |
| `raycast/**` | `just build-raycast` | reload the extension in Raycast |
| native app | from `mozeidon-native-app` / Homebrew | — |

Build-only/devDependency changes (e.g. a webpack bump) don't change runtime behaviour, so no
reload is needed for those.

**Relationship to upstream:** functionally **0 features behind** — the security fix, MV3 startup,
and multi-browser/profiles are all present (profiles is identical to upstream). By-SHA the
divergence reads "ahead/behind" only because the history was rebased to a 2024 merge-base, so the
cherry-pick well is effectively dry. Full detail + rationale: [`WORKSTATION_SETUP.md`](WORKSTATION_SETUP.md).

```bash
git fetch upstream
git log --oneline main..upstream/main     # check (rarely anything new)
git diff --stat main...upstream/main      # scope before porting
```

If something genuinely new appears upstream, **port it surgically** (reimplement) — never merge.
The rewritten `firefox-addon/` and the fork-only `tabs pick` / `tabs activate` features conflict
heavily with upstream's tree.

**Publishing a new Firefox version to AMO** (auto-updates every install):

```bash
# bump firefox-addon/manifest.json "version" first, then:
just submit-firefox     # build → package → web-ext sign --channel listed (+ source upload)
```

Needs AMO dev-hub JWT credentials in the environment — `WEB_EXT_API_KEY` (issuer) and
`WEB_EXT_API_SECRET` (secret), from <https://addons.mozilla.org/developers/addon/api/key/>.
Never put them in the repo or on the command line. Listed submissions go through AMO review
(source is uploaded for the bundled build); once approved, Firefox auto-updates installs.
