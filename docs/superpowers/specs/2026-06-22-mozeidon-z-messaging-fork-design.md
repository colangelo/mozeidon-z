# Fork & harden the native app as `mozeidon-z-messaging` — Design

**Date:** 2026-06-22
**Status:** approved (pending spec review)
**Repos touched:** new `colangelo/mozeidon-z-messaging` (the fork) + `colangelo/mozeidon-z` (the hub)

## Goal

Bring the last externally-owned runtime component of the Mozeidon-Z stack in-house: fork
`egovelox/mozeidon-native-app`, rebrand its binary to **`mozeidon-z-messaging`**, sign and
distribute it from our own Homebrew tap, add `--version`/`--help`, harden the code, and
**document how it works**. After this, the entire stack (CLI, both extensions, and the
native-messaging bridge) is forked, signed, and self-distributed — with no egovelox runtime
dependency.

This is a supply-chain / distribution + maintainability change, **not** a feature change. The
multi-browser/profiles capability already landed upstream (commit `0ea51b4`) and is included.

## Locked decisions

| # | Decision |
|---|---|
| 1 | New **GitHub-only** repo `colangelo/mozeidon-z-messaging` (no Gitea mirror, no Woodpecker). |
| 2 | Local checkout at **`~/_sync/dev/mozeidon-z-messaging`**, seeded from the existing clone's git history (keeps egovelox's 3 commits for provenance); egovelox remote demoted to `upstream`, `origin` → the new GitHub repo. The existing clone under `firefox-ai/mozeidon-native-app` is left untouched. |
| 3 | Binary renamed `mozeidon-native-app` → **`mozeidon-z-messaging`**. |
| 4 | **Host name `"mozeidon"` and IPC socket `mozeidon_native_app` are FROZEN** — unchanged, so **no AMO re-submit** and no break of the native-app⇄CLI contract. |
| 5 | Release via **goreleaser + GitHub Actions** on `v*` tag → cosign-signed public GitHub Release + auto-bumped `colangelo/homebrew-tap`. |
| 6 | Build on **latest stable Go (1.26.4)**; `go.mod` → `go 1.26`. |
| 7 | **MIT** license. |
| 8 | Platforms: **darwin (universal) + linux (amd64, arm64)**. Windows dropped. |

## The three names (why the rename is AMO-free)

| "Name" | Lives in | Depended on by | This change |
|---|---|---|---|
| **Host name** `"mozeidon"` | manifest `"name"` + extension `connectNative(ADDON_NAME)` (`firefox-addon/src/app.ts:28`) | the shipped AMO extension | **FROZEN** |
| **Binary filename** `mozeidon-native-app` | manifest `"path"`, brew formula, `depends_on`, `justfile:159`, docs | our repos only | → `mozeidon-z-messaging` |
| **IPC socket** `mozeidon_native_app` | native-app generates it; CLI reads it (`cli/core/app.go:25` fallback) | native-app + CLI (both ours) | **FROZEN** (no benefit to changing) |

## Code changes (in `mozeidon-z-messaging`)

Source today: `main.go` (140 LOC) + `models/registration-info.go` + `models/registered-native-app.go` (~228 LOC total).

1. **`--version` / `-v` / `--help` / `-h`.** At the top of `main()`: if `len(os.Args) > 1` and
   `os.Args[1]` is one of those flags, print and `os.Exit(0)`; otherwise fall through to
   `webBrowserProxy()`. Safe because the browser always launches the host with a manifest-path
   (+ extension-id) argument, never a flag. `--help` text notes it's a native-messaging host not
   meant to be run directly. Version: `var version = "dev"`, overridden at release time via
   goreleaser ldflags `-X main.version={{.Version}}`. `--version` prints `mozeidon-z-messaging <version>`.
2. **Panic guard** in `models/registered-native-app.go`: `response.Data.ProfileId[:8]` is used for
   both `IpcName` and `FileName`; guard `len(ProfileId) >= 8` (and ideally validate it parses as a
   UUID) → return a clean error instead of an index-out-of-range panic on malformed registration.
3. **Stop swallowing codec errors** in `main.go`'s loop: `json.Unmarshal(message.Data, …)` and
   `json.Marshal(response)` currently discard errors (`_`); return/log them (to **stderr** — stdout
   is the native-messaging channel and must never be written to outside the protocol).
4. **Robust end-of-stream.** Replace the exact-string compare `string(responseMessage) ==
   '{"data":"end"}'` with a parsed check on the unmarshaled response (`resp.Data == "end"`). One-sided
   and backward-compatible — the extension still sends the same `{"data":"end"}`.
5. **Tighten socket perms.** `ipc.ServerConfig.UnmaskPermissions: true → false` (native-app and CLI
   run as the same user on a single-user laptop, so a world-writable socket is unnecessary).
   **Gated on the end-to-end test** — revert if it breaks the local IPC.
6. **Cleanup.** Drop `-tags=pro dev` from goreleaser (no such build tags exist in the source); add
   `LICENSE` (MIT); `go.mod` → `go 1.26`; resolve/trim the stale `// TODO` comments (stderr logging
   is already correct, so the "log to a file" TODO is obsolete).

### Module path note
`go.mod` module is `github.com/egovelox/mozeidon-native-app`, and `main.go` imports
`github.com/egovelox/mozeidon-native-app/models`. Rename the module to
`github.com/colangelo/mozeidon-z-messaging` and update the `models` import accordingly.

## Release pipeline (goreleaser — approach ①)

- **`.goreleaser.yaml`:** `project_name: mozeidon-z-messaging`; builds darwin+linux × amd64/arm64,
  `CGO_ENABLED=0`, ldflags `-s -w -X main.version={{.Version}}`; `universal_binaries: { replace: true }`
  (darwin → one fat binary); `signs:` block = **cosign keyless** (`cosign sign-blob --yes`, OIDC);
  `brews:` → repository `colangelo/homebrew-tap`, formula `mozeidon-z-messaging`, with
  `test: system "#{bin}/mozeidon-z-messaging", "--version"`. Remove `-tags=pro dev`.
- **`.github/workflows/release.yml`:** trigger on `v*` (+ `workflow_dispatch`); `setup-go` `stable`;
  permissions `contents: write` + `id-token: write` (cosign); run `goreleaser release --clean`; env
  `GITHUB_TOKEN` (release) + `HOMEBREW_TAP_TOKEN` (tap push, the same secret pattern as the CLI).
- **`.github/workflows/ci.yml` (new):** on push/PR → `go vet ./... && go build ./... && go test ./...`.
  Closes the build-check gap upstream lacked.

## Documentation (in `mozeidon-z-messaging`) — "how everything works"

1. **README.md** (user-facing) — what it is; install (`brew install colangelo/tap/mozeidon-z-messaging`);
   the native-messaging manifest setup (`"name":"mozeidon"`, `"path"` → the new binary); `--version` /
   `--help`; build-from-source; releases. Credits the egovelox upstream.
2. **CLAUDE.md** (agent guidance) — build/test/release commands; architecture summary; the frozen
   names; pointer to ARCHITECTURE.md.
3. **ARCHITECTURE.md** (the core "how it works" doc) — covers:
   - The **native-messaging protocol**: 4-byte little-endian length-prefixed JSON over the host's
     stdin/stdout; why stdout is sacred (protocol channel).
   - The **IPC layer**: `james-barrow/golang-ipc` Unix-socket server; the message loop;
     request → browser → streamed responses → `{"data":"end"}` terminator.
   - The **registration handshake & profiles**: first message from the extension carries browser/
     profile metadata; the host writes a per-instance profile file to
     `$UserConfigDir/mozeidon_profiles/<pid>_<profileId8>.json` containing the per-instance socket
     name `mozeidon_native_app_<pid>_<profileId8>`.
   - The **3-way contract**: extension (`registration.ts`) → native-app (`models/`) → CLI
     (`cli/profiles/profiles.go`), and the legacy single-socket fallback `mozeidon_native_app`.
   - **Lifecycle**: SIGTERM/SIGINT + `defer` unregister (removes the profile file); the Windows gap.
   - **Security notes**: socket permissions, golang-ipc's homegrew encryption handshake (not audited),
     localhost-only trust model.
   - An ASCII data-flow diagram.

## Hub changes (in `colangelo/mozeidon-z`)

- `justfile:159` — manifest `"path"` → `/opt/homebrew/bin/mozeidon-z-messaging`.
- `.github/workflows/release.yml` — the base64 formula template (`FORMULA_B64`): change
  `depends_on "egovelox/mozeidon/mozeidon-native-app"` → `depends_on "colangelo/mozeidon-z-messaging"`
  (i.e. our tap formula). Re-encode the base64.
- Docs — `README.md`, `WORKSTATION_SETUP.md`, `CLAUDE.md`, `CI_RELEASE_RUNBOOK.md`,
  `ACTIVATE_TAB_IMPLEMENTATION.md`: replace `mozeidon-native-app` with `mozeidon-z-messaging` where it
  refers to *our* bridge; keep upstream references in "relationship to upstream" prose. Add a
  `2026-06-22` audit-log entry in `WORKSTATION_SETUP.md`.

## Testing

- **Unit (Go, new in the fork):** `models` profile generation (filename/ipcName formatting; the new
  length guard rejects a short/empty `ProfileId`); flag parsing (`--version`, `--help`, and that a
  non-flag first arg falls through to the proxy path).
- **End-to-end (manual, gated):** build → install the renamed binary → `just setup-native-messaging`
  (new path) → restart Firefox → `mozeidon-z tabs get` streams open tabs through the renamed bridge;
  `pgrep -fl mozeidon-z-messaging` shows it; **confirm the `UnmaskPermissions: false` change did not
  break IPC** (revert that one change if it did).

## Rollout / migration

1. Fork's first release **`v1.0.0`** (default — a clean version line for the freshly-named binary;
   the bridge's version is independent of the CLI's 5.x) → formula lands in `colangelo/homebrew-tap`.
2. `brew uninstall egovelox/mozeidon/mozeidon-native-app` → `brew install colangelo/tap/mozeidon-z-messaging`.
3. `just setup-native-messaging` (writes the new `"path"`) → restart Firefox.
4. Because the host name `"mozeidon"` is unchanged, the extension reconnects with zero changes.
5. `mozeidon-z` CLI's `depends_on` now pulls our bridge, so fresh installs get it automatically.

## Scope cuts (YAGNI)

- **Windows build** dropped — macOS-focused stack; the SIGTERM-based unregister doesn't work on
  Windows anyway.
- **Windows stale-profile-file sweep** (a CLI-side change) deferred — moot once Windows isn't shipped.
- **Renaming the IPC socket** — explicitly not done (no benefit; CLI hardcodes the fallback).

## Risks / things to watch

- **`UnmaskPermissions: false`** — primary behavioral risk; covered by the e2e gate above.
- **Version line at first tag** — defaulting to **`v1.0.0`** (clean line for the renamed binary).
  Overridable at tag time if we'd rather mirror egovelox 4.x or the `mozeidon-z` 5.x line.
- **License provenance** — upstream repo has no LICENSE file (`licenseInfo: null` on GitHub) though
  it's public; we add MIT and credit egovelox. (Egovelox's own CLI/formula are MIT-declared.)
- **`HOMEBREW_TAP_TOKEN`** must exist on the new repo too (same PAT, added as a secret on
  `colangelo/mozeidon-z-messaging`).
