# Workstation Setup & Toolchain Audit

How the **Mozeidon-Z** browser-tab toolset is installed and wired on the Apple Silicon
workstations (`m4m` / `ac-mbm5`), what is built-here vs. Homebrew, and the record of the
cleanups. This is the living source of truth for *how the moving parts fit together*;
per-component detail lives in each component's own README/CLAUDE.

> This repo is **`colangelo/mozeidon-z`** — originally a fork of `egovelox/mozeidon`, now a
> standalone hard fork (**Mozeidon-Z**, see *Relationship to upstream* below). "The project"
> below always means this repo.

---

## Live pipeline (what actually runs)

```
Firefox extension          native-app (Homebrew)              CLI (Homebrew tap)        front-end
mozeidon-z@a-layer.io  ──►  /opt/homebrew/bin/            ──►  /opt/homebrew/bin/     ──►  Raycast
(this repo's firefox-addon)  mozeidon-z-messaging               mozeidon-z 5.0.2           (raycast/ ext)
   v5.0.0                    (native-messaging bridge)          (colangelo/tap)
```

- The browser talks to a **native-messaging host** declared at
  `~/Library/Application Support/Mozilla/NativeMessagingHosts/mozeidon.json`, which points
  at the Homebrew `mozeidon-z-messaging`. Allowed extensions:
  `mozeidon-z@a-layer.io`, `mozeidon@anthropic.github.io`, `mozeidon-dev@ac.local`
  (written by `just setup-native-messaging`).
- The native app shells out to the **CLI on `PATH`**, which is the Homebrew-tap binary
  `mozeidon-z` at `/opt/homebrew/bin/mozeidon-z` (`brew install colangelo/tap/mozeidon-z`).
  For local CLI work, `just install-cli` builds a dev `mozeidon-z` into `~/.local/bin/` instead.
- **Raycast** is the primary front-end (extension under `raycast/`, dev-installed). It
  calls the same CLI.

End-to-end smoke test:

```bash
mozeidon-z --version     # → mozeidon-z version 5.0.2 (Homebrew tap)
mozeidon-z tabs get      # streams JSON of open Firefox tabs (needs Firefox open w/ ext)
pgrep -fl mozeidon-z-messaging   # bridge process, spawned by Firefox
```

---

## Component inventory

| Component | Source of truth | Where it lives | Version | Notes |
|---|---|---|---|---|
| **CLI** | **here** (`cli/`); shipped via **Homebrew tap** `colangelo/tap` | `/opt/homebrew/bin/mozeidon-z` (on `PATH`) | 5.0.2 | `brew install colangelo/tap/mozeidon-z` — released by GitHub Actions (see [`CI_RELEASE_RUNBOOK.md`](CI_RELEASE_RUNBOOK.md)). Dev builds: `just install-cli` → `~/.local/bin/mozeidon-z`. |
| **Firefox extension** | **here** (`firefox-addon/`) | loaded in Firefox as `mozeidon-z@a-layer.io` | 5.0.0 | Built `.xpi` is a gitignored artifact; release via `just submit-firefox` (AMO, auto-updates installs). |
| **Chrome extension** | **here** (`chrome-addon/`) | not loaded | 5.0.0 | `chrome-addon/src/` is generated from `firefox-addon/src/` (verbatim copy by `just build-chrome`) and gitignored. Kept in sync for completeness; not in active use. |
| **native-app** | **fork** `colangelo/mozeidon-z-messaging` (own repo), shipped via `colangelo/tap` | `/opt/homebrew/bin/mozeidon-z-messaging` | 1.0.0 | The browser bridge. Built in its own repo; pulled automatically by `brew install colangelo/tap/mozeidon-z` via `depends_on`. |
| **Raycast extension** | **here** (`raycast/`) | Raycast dev extension | — | Primary front-end. Raycast handles its own versioning (no semver in `package.json`). |

### Not built here, and why it doesn't matter

- **`mozeidon-z-messaging`** — the bridge, now a **hard fork** of `egovelox/mozeidon-native-app`
  in its own repo (`colangelo/mozeidon-z-messaging`), shipped via `colangelo/homebrew-tap` and
  pulled automatically by `brew install colangelo/tap/mozeidon-z` via `depends_on`. We added
  `--version`/`--help` + hardening but keep its **wire protocol frozen** (host name `mozeidon`,
  IPC socket `mozeidon_native_app`), so it stays a drop-in transparent `{command, args}` proxy.
- **`mozeidon-macos-ui`** — a stock egovelox Swift menu-bar app (a Spotlight-style
  alternative front-end). **Never built here, never used** — superseded by Raycast. Removed
  2026-06-14 (see audit).

---

## Relationship to upstream (`egovelox/mozeidon`)

**Mozeidon-Z is a hard fork.** Treat it as a standalone project, not a living branch to keep
merged with upstream.

The numbers look dramatic but are misleading. By SHA the divergence is **117 ahead / 59
behind**, but the merge-base is `1709f4a` from **2024-03-23** — the fork's history was rebased
to a linear, fork-only line that no longer shares upstream's individual commits. So those "59
behind" are almost entirely old upstream commits (history, groups, bookmarks, 3.0.0 release,
READMEs) that this project has **reimplemented**, not genuinely missing.

**Functionally we are 0 features behind upstream.** The only upstream commits that postdate the
fork are these four, and all are accounted for:

| Upstream commit | What | Status here |
|---|---|---|
| `36cd5bb` | Fix npm-dependency vulnerabilities | ✅ **reimplemented** (webpack bump + `npm audit`, 2026-06-14) |
| `0cd8f39` | Improve browser-extension startup for Manifest V3 | ✅ **ported** (2026-06-15) |
| `05663ab` | Supporting multi browsers / profiles | ✅ **already present** — every profiles file (`cli/cmd/profiles/*`, `cli/profiles/profiles.go`, `firefox-addon/src/services/profiles.ts`, `registration.ts`, …) exists here **identically** to upstream |
| `4b8d906` | Bump browser-extension to 4.1 + README | ➖ **N/A** — upstream's own version line; we run our own (now 5.0.0) |

**The upstream cherry-pick well is dry.** Going forward, scan occasionally but expect nothing:

```bash
git fetch upstream
git log --oneline main..upstream/main          # what (if anything) is new upstream
git diff --stat main...upstream/main           # scope before porting
```

If something genuinely new appears upstream, port it **surgically** (reimplement), never merge
— the rewritten `firefox-addon/` + the fork's `tabs pick` / `tabs activate` features conflict
heavily with upstream's tree.

### What only exists here (would be lost in any "re-base on upstream")

- **`tabs pick`** — interactive fuzzy tab picker (`cli/core/tabs-pick.go` + `cmd/tabs/pick.go`)
- **`tabs activate`** — AppleScript window-to-front focusing, across macOS Spaces
  (`cli/core/tabs-activate.go` + `cmd/tabs/activate-tab.go` + the addon `activateTab`)
- The Raycast focus fix, `--version` / `install-cli`, `just submit-firefox` (AMO), branding
  (mazinger-Z icons, Mozeidon-Z), the rewritten README/CLI_REFERENCE, and the agent tooling.

A literal `git rebase upstream/main` would replay 117 commits across a 2024 merge-base — hours
of conflict resolution for ~zero gain. Don't.

---

## Maintenance

| Task | Command |
|---|---|
| Install / upgrade the released CLI | `brew install colangelo/tap/mozeidon-z` (or `brew upgrade mozeidon-z`) |
| Rebuild + install the CLI (dev) | `just install-cli` (→ `~/.local/bin/mozeidon-z`) |
| Cut a CLI release | bump `cli/cmd/root.go` `var Version`, tag `vX.Y.Z`, `git push origin vX.Y.Z` — see [`CI_RELEASE_RUNBOOK.md`](CI_RELEASE_RUNBOOK.md) |
| Rebuild the Firefox bundle | `just build-firefox` |
| Release a new Firefox version to AMO | `just submit-firefox` (bump `firefox-addon/manifest.json` first; auto-updates installs) |
| Rebuild everything | `just build-all` |
| Re-audit extension deps | `cd firefox-addon && npm audit` (and `chrome-addon`) |
| Check the native-messaging wiring | `just check-native-messaging` |

### PATH-shadow gotcha

The CLI you actually run is the **Homebrew-tap build** `mozeidon-z` at `/opt/homebrew/bin/mozeidon-z`
(`brew install colangelo/tap/mozeidon-z`) — that's the source of truth now. `just install-cli` builds a
*dev* `mozeidon-z` into `~/.local/bin/`; if `~/.local/bin` precedes `/opt/homebrew/bin` on `PATH`, that
dev build **shadows** brew's. So after a release, `brew upgrade mozeidon-z` and remove any stale
`~/.local/bin/mozeidon-z` (or knowingly let the dev build win). Do **not** `brew install
egovelox/mozeidon/mozeidon` — that's the upstream-named CLI (`mozeidon`, no `-z`), a different binary
unrelated to this fork. Only the **native-app** belongs to the egovelox tap.

---

## Audit log

### 2026-06-22 — native-app forked + renamed to `mozeidon-z-messaging`

**Context**
- The native-messaging bridge was previously sourced from `egovelox/mozeidon-native-app` via the
  `egovelox/mozeidon` Homebrew tap. To give the project full control over the bridge (versioning,
  security fixes, distribution), the native-app was forked into `colangelo/mozeidon-z-messaging`
  and published to `colangelo/homebrew-tap`.

**Actions**
1. **Bridge renamed** `mozeidon-native-app` → `mozeidon-z-messaging` (binary + formula). The IPC
   socket name (`mozeidon_native_app`) and native-messaging host name (`"mozeidon"`) are unchanged —
   no AMO or extension change required.
2. **CLI formula `depends_on`** updated from `"egovelox/mozeidon/mozeidon-native-app"` to
   `"colangelo/mozeidon-z-messaging"` (in `.github/workflows/release.yml` `FORMULA_B64`).
3. **`just setup-native-messaging`** manifest `"path"` updated to `/opt/homebrew/bin/mozeidon-z-messaging`.
4. **Docs** — live-pipeline diagram, component inventory, and prose updated throughout to reference
   `mozeidon-z-messaging` (shipped via `colangelo/homebrew-tap`). Historical audit entries preserved as-is.

### 2026-06-22 — public distribution via Homebrew tap + CLI rename

**Findings**
- The CLI shipped only as a local `just install-cli` build (`~/.local/bin/mozeidon`) — no one-command
  install for a fresh machine — and Woodpecker's old release steps published to tailnet-only Gitea
  Releases, which aren't publicly installable.

**Actions**
1. **Renamed the CLI command/binary** `mozeidon` → `mozeidon-z` (cosmetic; the native-messaging
   plumbing stays `mozeidon` / `mozeidon_native_app`, so **no AMO change**). Bumped CLI to **5.0.2**.
2. **Moved releases to GitHub Actions** (`.github/workflows/release.yml`, HittyPing pattern): a `v*`
   tag on GitHub `origin` → cross-compile (darwin/linux × arm64/amd64) → cosign-sign → **public
   GitHub Release** → auto-bump the formula in **`colangelo/homebrew-tap`** (needs the
   `HOMEBREW_TAP_TOKEN` secret). Woodpecker is now **build-CI only** (release steps removed).
3. **Install is now** `brew install colangelo/tap/mozeidon-z` (also pulls the `mozeidon-z-messaging`
   bridge via `depends_on`). Removed the orphaned `~/.local/bin/mozeidon` (old 4.1.1 build).
4. **Docs** — rewrote [`CI_RELEASE_RUNBOOK.md`](CI_RELEASE_RUNBOOK.md) for the GitHub Actions +
   Homebrew flow (one-time prereqs, the `HOMEBREW_TAP_TOKEN` secret, the `gh --repo` gotcha, and the
   rerun-on-failure fix for a red Homebrew step), and updated this file's live-pipeline / inventory /
   maintenance / PATH-shadow sections.

### 2026-06-16 — Mozeidon-Z 5.0.0 + housekeeping

**Findings**
- Re-compared against upstream after the history rebase: by-SHA 117/59, but merge-base is
  2024 (`1709f4a`). **Functionally 0 features behind** — including multi-browser/profiles
  (`05663ab`), whose files are present and identical here. The earlier "profiles deferred"
  note was stale.
- Build artifacts were tracked in git: stale `source-v*.zip` bundles in both addons, a 9.4M
  CLI binary and a broken self-referential `Mozeidon-Z` symlink at the repo root, plus empty
  claude-mem `CLAUDE.md` placeholders.

**Actions**
1. **Raycast focus fix** — `Open Tab` now closes the Raycast window *before* activating
   Firefox, so the target window reliably comes to the front (was a `closeMainWindow` race).
2. **Ported upstream MV3 startup** (`0cd8f39`) into `firefox-addon/src/app.ts`
   (`onStartup`/`onInstalled` + `connected` guard + reconnect); chrome-addon inherits it.
3. **Removed beads** issue tracker entirely (`.beads/`, hooks, docs) — unused, and its
   pre-commit hook was failing on a repo-ID mismatch and blocking commits.
4. **gitignore hygiene** — untracked the stale source zips, root binary, broken symlink, and
   claude-mem placeholders; generalized the package patterns.
5. **Declared the hard fork**: bumped CLI + both addons to **5.0.0**, finalized the
   **Mozeidon-Z** name, and recorded that the upstream cherry-pick well is dry.

### 2026-06-14 — cleanup + security sync

**Findings**
- The whole pipeline was live and healthy; the CLI on `PATH` was the fork build.
- Homebrew had three formulae: `mozeidon` (CLI), `mozeidon-native-app`, `mozeidon-macos-ui`.
- `mozeidon-macos-ui` was stock egovelox, never used (no `LaunchAgent`, never launched) — a
  redundant alternative to Raycast, and the only reason the shadowed brew `mozeidon` CLI
  stayed installed.
- The fork was missing an upstream npm-dependency **security fix** (`36cd5bb`): a build-time
  webpack SSRF (GHSA-8fgc-7cc6-rx7x, GHSA-38r7-794h-5758) plus a `serialize-javascript`/
  `terser-webpack-plugin` vuln. Build-time only (webpack is a devDependency) — no runtime
  exposure in the shipped extension.

**Actions**
1. **Brew cleanup** — removed the never-used UI cask and the redundant shadowed CLI; kept the
   `mozeidon-native-app` bridge. Shadow warning gone; pipeline re-verified.
2. **Security sync** — reproduced the fix surgically (bumped `webpack` `5.103.0 → 5.106.1` in
   both addons, regenerated lockfiles, `npm audit fix`): 6 vulnerabilities → 0.
3. **Documentation** — created this file.
