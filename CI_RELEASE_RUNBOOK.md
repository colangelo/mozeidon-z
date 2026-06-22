# CI & Release Runbook — Build (Woodpecker) + Release (GitHub Actions → Releases + Homebrew)

The worked, reusable runbook for this repo (`A-Layer/mozeidon-z` on Gitea, `colangelo/mozeidon-z`
on GitHub). **Two systems, two jobs:**

| Job | System | Trigger | Produces |
|---|---|---|---|
| **Build CI** | Woodpecker (`ci.cat-bluegill.ts.net`, `.woodpecker.yml`) | push / PR / manual | green build of CLI + both add-ons (mirrors `just build-all`) |
| **Release** | GitHub Actions (`.github/workflows/release.yml`) | `v*` tag on GitHub `origin` | public **GitHub Release** (signed binaries) + **Homebrew** formula bump |

> **Relationship to the canonical standard.** The repo-agnostic homelab CI/CD pattern lives in the
> **home-network** repo at `docs/ci-release-standard.md` (Gitea + Woodpecker + **Gitea Releases**).
> mozeidon-z follows that standard for **build CI** but **diverges for releases**: it ships to a
> **public** audience via `brew install`, and Gitea Releases are tailnet-only. So releases moved to
> GitHub Actions + a public Homebrew tap (the "HittyPing pattern"). Keep the build half in sync with
> the standard; the release half is a deliberate, mozeidon-specific delta. **[repo-specific]**

## The stack & the two remotes

- **Gitea** `gitea.cat-bluegill.ts.net` — git hosting (tailnet-only). `internal` remote. Its webhook
  drives Woodpecker.
- **Woodpecker** `ci.cat-bluegill.ts.net` — **build CI only** (no release steps; they were removed
  when releases moved to GitHub).
- **GitHub** `github.com/colangelo/mozeidon-z` — **public** `origin` remote. Hosts GitHub Actions +
  the public GitHub Releases that `brew` downloads from.
- **Homebrew tap** `github.com/colangelo/homebrew-tap` (public) — `brew install colangelo/tap/mozeidon-z`.
- Browser extensions ship to **AMO** / Chrome Web Store separately, *not* via CI. **[repo-specific]**

**Where pushes go — get this right or releases misfire:**

| You push… | …to `internal` (Gitea) | …to `origin` (GitHub) |
|---|---|---|
| `main` | ✅ Woodpecker build runs | ✅ no trigger (Actions only run on tags) |
| `v*` tag | ❌ **never** — Gitea Actions must stay disabled | ✅ **yes** — fires the release workflow |

---

# Part A — Build CI (Woodpecker)

### Auth — `woodpecker-cli` (watch / debug / trigger)

```bash
export WOODPECKER_SERVER=https://ci.cat-bluegill.ts.net
export WOODPECKER_TOKEN=$(op read 'op://AC-DevOps/woodpecker - Personal Access Token/password')
woodpecker-cli info        # → User: ac
```

> **1Password / AFK caveat:** `op read` triggers Touch ID. If you're away the prompt **times out** —
> that is **not** a failure. Don't loop-retry; wait and re-run. Probe readiness with `op vault list`,
> never `op whoami`.
>
> **Vault note:** the Woodpecker PAT is in the **AC-DevOps** vault. Some repos' `.envrc` reference
> `op://Private/...` — that's a different account context; on this workstation it's AC-DevOps.

### One-time setup

1. **Activate the repo in Woodpecker** — `ci.cat-bluegill.ts.net` → enable `A-Layer/mozeidon-z`. This
   installs the Gitea webhook; until then pushes run nothing. Ensure a Woodpecker **agent** is online.
   (No *Trusted* flag needed — a plain build pipeline doesn't touch `docker.sock`.)
2. **No release secret needed here anymore.** The old `gitea_token` secret existed only for the
   retired Gitea-Releases pipeline; it's an orphan now (delete it — see *Loose ends*).

### The pipeline (`.woodpecker.yml`)

Build-only. One step per artifact (`event: [push, pull_request, manual]`), each with `depends_on: []`
so they run in parallel — mirrors `just build-all`:

- CLI (`golang:1.24`) — `cd cli && go build`
- Firefox add-on (`node:20`) — `npm ci && npm run build` in `firefox-addon/`
- Chrome add-on (`node:20`) — sync from firefox, then `npm ci && npm run build`

### Monitoring & debugging

```bash
woodpecker-cli pipeline ls   A-Layer/mozeidon-z                  # recent runs
woodpecker-cli pipeline last A-Layer/mozeidon-z                  # latest
woodpecker-cli pipeline ps   A-Layer/mozeidon-z <N>              # step states for run N
woodpecker-cli pipeline log show A-Layer/mozeidon-z <N> <step>   # tail a step's log
woodpecker-cli pipeline start  A-Layer/mozeidon-z <N>            # re-run run N (same commit)
woodpecker-cli pipeline create A-Layer/mozeidon-z --branch main  # fresh manual run
```

A pipeline in status **`error` with zero steps** = a config-compile failure; read it from the API:

```bash
RID=$(curl -s -H "Authorization: Bearer $WOODPECKER_TOKEN" "$WOODPECKER_SERVER/api/repos/lookup/A-Layer/mozeidon-z" | jq -r .id)
curl -s -H "Authorization: Bearer $WOODPECKER_TOKEN" "$WOODPECKER_SERVER/api/repos/$RID/pipelines/<N>" | jq '{status, errors}'
```

### Woodpecker gotchas (learned during bring-up)

1. **Woodpecker expands `${...}` across the WHOLE config — including comments** — and errors
   (`unable to parse variable name`) on anything it can't parse, compiling to **zero steps** (status
   `error`). Keep shell param-expansions (`${t%/*}`) and literal `${...}` placeholders OUT of
   `.woodpecker.yml`. Use `$VAR`, `$(...)`, or literals; only `${KNOWN_CI_VAR}` forms it supports are safe.
2. **Steps go parallel under a DAG once any step has `depends_on`.** Give every step an explicit
   `depends_on` (`[]` = independent) so ordering is unambiguous.

---

# Part B — Release (GitHub Actions → GitHub Releases + Homebrew)

### Auth & the GitHub CLI gotcha

```bash
# ⚠️ This repo has THREE remotes incl. the upstream fork source. `gh` resolves to
#    egovelox/mozeidon (upstream) by default — ALWAYS pass --repo for our fork:
command gh run list   --repo colangelo/mozeidon-z
command gh release view v5.0.2 --repo colangelo/mozeidon-z
# (or pin it once:  gh repo set-default colangelo/mozeidon-z)
```

- The workflow authenticates to GitHub Releases with the **auto-provided `GITHUB_TOKEN`** (no setup).
- The Homebrew bump needs a **`HOMEBREW_TAP_TOKEN`** repo secret = a PAT with **write to
  `colangelo/homebrew-tap`** (reuse HittyPing's, or a fine-grained PAT scoped to that one repo).

### One-time prerequisites (all currently ✅, listed so a fresh clone / new maintainer can re-verify)

| # | Prereq | Why | Check |
|---|---|---|---|
| 1 | `colangelo/mozeidon-z` **public** | `brew` downloads release assets unauthenticated | `gh repo view colangelo/mozeidon-z --json visibility` |
| 2 | `colangelo/homebrew-tap` exists & **public** | the formula lives there; `brew install colangelo/tap/…` reads it | `gh repo view colangelo/homebrew-tap --json visibility` |
| 3 | **`HOMEBREW_TAP_TOKEN`** secret set on `mozeidon-z` | the `update-homebrew` job pushes the formula bump | `gh secret list --repo colangelo/mozeidon-z` |
| 4 | **Gitea Actions DISABLED** for `A-Layer/mozeidon-z` | else the mirror also runs `release.yml` on the tag and fails | Gitea → repo → Settings → Actions |

```bash
# Set/rotate the tap token (value lives in 1Password — never on the command line as plaintext):
command gh secret set HOMEBREW_TAP_TOKEN --repo colangelo/mozeidon-z \
  --body "$(op read 'op://Private/<tap-pat-item>/token')"
```

### What `release.yml` does — 4 chained jobs

`v*` tag (or `workflow_dispatch` with a `tag` input) →

1. **test** — checkout the tag, `go build ./... && go test ./...`.
2. **build** (matrix: darwin/linux × arm64/amd64) — `CGO_ENABLED=0 go build -ldflags="-s -w"`,
   upload each binary as an artifact (`retention-days: 1`).
3. **release** — download all binaries, `sha256sum` → `checksums.txt`, **cosign keyless sign**
   (`sign-blob`, OIDC `id-token: write`) each binary + the checksums (`.sig` + `.pem` per file),
   then publish a **public GitHub Release** with `generate_release_notes: true`.
4. **update-homebrew** — `curl` each published binary to recompute its sha256, render the embedded
   formula template (base64 in the workflow), substitute version + the 4 sha256s, then
   clone → commit → **push to `colangelo/homebrew-tap`** using `HOMEBREW_TAP_TOKEN`.

The formula `depends_on "colangelo/mozeidon-z-messaging"`, so `brew install` of the CLI also
pulls the native-messaging bridge.

### Cutting a release

```bash
# 1. bump the version the binary reports  [repo-specific: cli/cmd/root.go `var Version`], commit
# 2. tag and push to ORIGIN (GitHub) ONLY:
git tag -a vX.Y.Z -m "Mozeidon-Z X.Y.Z"
git push origin vX.Y.Z
# (push main to both remotes as usual; push the TAG to origin only — see the remotes table above)
```

The tag fires `release.yml` → signed binaries on the public GitHub Release → formula bumped in the tap.
Install / upgrade:

```bash
brew install colangelo/tap/mozeidon-z      # or: brew upgrade mozeidon-z
```

### Verify a release

```bash
command gh run list    --repo colangelo/mozeidon-z --limit 5
command gh release view vX.Y.Z --repo colangelo/mozeidon-z \
  --json tagName,isDraft,assets --jq '{tag:.tagName,draft:.isDraft,assets:[.assets[].name]}'
# tap landed?  (look for the "Update mozeidon-z to vX.Y.Z" commit and substituted version/sha256)
command gh api repos/colangelo/homebrew-tap/commits --jq '.[0].commit.message'
```

### When the **Homebrew** job fails (the classic exit 128)

Symptom — `release`/`build` all green but **Update Homebrew Formula** fails:

```
remote: Invalid username or token. Password authentication is not supported for Git operations.
fatal: Authentication failed for 'https://github.com/colangelo/homebrew-tap.git/'   → exit 128
```

Cause: `HOMEBREW_TAP_TOKEN` is **missing or expired**, so `GH_TOKEN` is empty and the `git push`
can't authenticate. The GitHub Release itself already published fine (it uses `GITHUB_TOKEN`).

**Fix without re-cutting the release** — the signed binaries already exist, so just fix the token and
re-run the failed job; its checksum step re-reads the published assets:

```bash
command gh secret set HOMEBREW_TAP_TOKEN --repo colangelo/mozeidon-z --body "$(op read 'op://Private/<tap-pat-item>/token')"
command gh run rerun <run-id> --failed --repo colangelo/mozeidon-z     # re-runs only the failed job
```

If the build artifacts have aged out (`retention-days: 1`) the rerun's checksum step still works (it
`curl`s the published Release assets, not the artifacts). Last-resort fallback: re-trigger the whole
thing via `workflow_dispatch` with the tag, or hand-commit the formula to the tap.

### GitHub / Homebrew gotchas

1. **`gh` defaults to the upstream fork** (`egovelox/mozeidon`) because it's a remote here — every
   `gh` command needs `--repo colangelo/mozeidon-z` (or `gh repo set-default`).
2. **Tags to `origin`, never `internal`.** A `v*` tag on Gitea with Gitea Actions enabled would run
   `release.yml` on the mirror and fail. Keep Gitea Actions disabled for this repo.
3. **The Homebrew bump is the only step that needs a secret.** A green Release with a red
   "Update Homebrew Formula" = the token, every time. The release is still usable from the GitHub
   Release page; only `brew` is behind until the formula lands.

---

## Loose ends / cleanup

- **Delete the orphaned `gitea_token` Woodpecker secret** (only the retired Gitea-Releases pipeline
  used it): `woodpecker-cli repo secret rm A-Layer/mozeidon-z gitea_token`.
- The **v5.0.0 / v5.0.1 Gitea releases & tags** (from the old approach) can be deleted on Gitea if you
  want a clean release list there.
