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
Firefox extension          native-app (Homebrew)              CLI                       front-end
mozeidon-z@a-layer.io  ──►  /opt/homebrew/bin/            ──►  ~/.local/bin/mozeidon  ──►  Raycast
(this repo's firefox-addon)  mozeidon-native-app 4.0.0          5.0.0 (built from cli/)     (raycast/ ext)
   v5.0.0                    (native-messaging bridge)          (this repo's build)
```

- The browser talks to a **native-messaging host** declared at
  `~/Library/Application Support/Mozilla/NativeMessagingHosts/mozeidon.json`, which points
  at the Homebrew `mozeidon-native-app`. Allowed extensions:
  `mozeidon-z@a-layer.io`, `mozeidon@anthropic.github.io`, `mozeidon-dev@ac.local`
  (written by `just setup-native-messaging`).
- The native app shells out to the **CLI on `PATH`**, which is the locally-built fork
  binary at `~/.local/bin/mozeidon` (built via `just install-cli`).
- **Raycast** is the primary front-end (extension under `raycast/`, dev-installed). It
  calls the same CLI.

End-to-end smoke test:

```bash
mozeidon --version       # → mozeidon version 5.0.0 (fork build)
mozeidon tabs get        # streams JSON of open Firefox tabs (needs Firefox open w/ ext)
pgrep -fl mozeidon-native-app   # bridge process, spawned by Firefox
```

---

## Component inventory

| Component | Source of truth | Where it lives | Version | Notes |
|---|---|---|---|---|
| **CLI** | **here** (`cli/`) | `~/.local/bin/mozeidon` (on `PATH`) | 5.0.0 | Built locally via `just install-cli`. This is the one that runs. |
| **Firefox extension** | **here** (`firefox-addon/`) | loaded in Firefox as `mozeidon-z@a-layer.io` | 5.0.0 | Built `.xpi` is a gitignored artifact; release via `just submit-firefox` (AMO, auto-updates installs). |
| **Chrome extension** | **here** (`chrome-addon/`) | not loaded | 5.0.0 | `chrome-addon/src/` is generated from `firefox-addon/src/` (verbatim copy by `just build-chrome`) and gitignored. Kept in sync for completeness; not in active use. |
| **native-app** | **Homebrew** `egovelox/mozeidon` tap | `/opt/homebrew/bin/mozeidon-native-app` | 4.0.0 | **Not built here.** The actual browser bridge. Installed *on request* (a `brew leaves` leaf). Keep it. |
| **Raycast extension** | **here** (`raycast/`) | Raycast dev extension | — | Primary front-end. Raycast handles its own versioning (no semver in `package.json`). |

### Not built here, and why it doesn't matter

- **`mozeidon-native-app`** — we use the stock Homebrew build. The project does not modify it;
  it is a transparent `{command, args}` proxy. The only upstream divergence that ever touched
  it is multi-browser/profile support, which the native-app 4.0.0 release already covers.
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
| Rebuild + install the CLI | `just install-cli` (→ `~/.local/bin/mozeidon`) |
| Rebuild the Firefox bundle | `just build-firefox` |
| Release a new Firefox version to AMO | `just submit-firefox` (bump `firefox-addon/manifest.json` first; auto-updates installs) |
| Rebuild everything | `just build-all` |
| Re-audit extension deps | `cd firefox-addon && npm audit` (and `chrome-addon`) |
| Check the native-messaging wiring | `just check-native-messaging` |

### PATH-shadow gotcha

The CLI you actually run is the **local build** at `~/.local/bin/mozeidon`. If you ever
`brew install egovelox/mozeidon/mozeidon`, Homebrew installs a *second* CLI at
`/opt/homebrew/bin/mozeidon` that is **shadowed** by the local build (and triggers a brew
"shadowed by" warning). Don't install the brew CLI formula — the fork build is the source
of truth. Only the **native-app** belongs to Homebrew.

---

## Audit log

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
