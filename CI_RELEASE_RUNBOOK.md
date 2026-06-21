# CI & Release Runbook — Gitea + Woodpecker + Gitea Releases

The worked, reusable runbook for the homelab CI/CD pattern, as implemented in this repo
(`A-Layer/mozeidon-z`). Written to be **portable**: most of it applies to any repo on the
`gitea.cat-bluegill.ts.net` (git) + `ci.cat-bluegill.ts.net` (Woodpecker) stack. Lines marked
**[repo-specific]** change per project.

## The stack

- **Gitea** `gitea.cat-bluegill.ts.net` — git hosting (tailnet-only); orgs owned by `ac`.
- **Woodpecker** `ci.cat-bluegill.ts.net` — CI, OAuth-linked to Gitea, reads `.woodpecker.yml`.
- **Artifacts** — CLI binaries → **Gitea Releases** (here); container images → **Harbor** (e.g. `direction`).
- Browser extensions ship to **AMO** / Chrome Web Store separately, *not* via CI. **[repo-specific]**

## Auth & access — two tokens, two stores (don't mix them up)

| Purpose | Token | Where it comes from |
|---|---|---|
| `woodpecker-cli` — watch / debug / trigger pipelines | Woodpecker PAT | `op read 'op://AC-DevOps/woodpecker - Personal Access Token/password'` |
| Gitea REST API **from the workstation** (mint tokens, manage releases) | Gitea PAT (`ac`) | macOS Keychain via `git credential fill` (no Touch ID) |
| Gitea REST API **inside a pipeline** (create release) | dedicated `gitea_token` secret | Woodpecker repo secret = a least-priv Gitea PAT |

```bash
# Woodpecker CLI
export WOODPECKER_SERVER=https://ci.cat-bluegill.ts.net
export WOODPECKER_TOKEN=$(op read 'op://AC-DevOps/woodpecker - Personal Access Token/password')
woodpecker-cli info        # → User: ac

# Gitea REST API from the workstation (keychain; no op/Touch-ID)
GITEA_TOKEN=$(printf 'protocol=https\nhost=gitea.cat-bluegill.ts.net\n\n' \
  | git credential fill | awk -F= '/^password/{print $2}')
```

> **1Password / AFK caveat:** `op read` triggers Touch ID. If you're away the prompt **times out** —
> that is **not** a failure. Don't loop-retry; wait and re-run. Probe readiness with `op vault list`,
> never `op whoami`. (The Gitea keychain path needs no Touch ID.)
>
> **Vault note:** the Woodpecker PAT is in the **AC-DevOps** vault. Some repos' `.envrc` reference
> `op://Private/...` — that's a different account context; on this workstation it's AC-DevOps.

## One-time per-repo setup

1. **Activate the repo in Woodpecker** — `ci.cat-bluegill.ts.net` → enable `<org>/<repo>`. This
   installs the Gitea webhook; until then pushes run nothing. Ensure a Woodpecker **agent** is online.
   (Mark the repo *Trusted* only if it needs `docker.sock`/service containers — a plain build/release
   pipeline like this one does not.)
2. **Mint a least-priv Gitea PAT for releases.** Gitea PATs are *user-scoped*; since `ac` owns all
   orgs, `write:repository` covers every repo. Token creation requires **basic auth** (not a token):
   ```bash
   GUSER=ac
   KC=$(printf 'protocol=https\nhost=gitea.cat-bluegill.ts.net\n\n' | git credential fill | awk -F= '/^password/{print $2}')
   curl -s -u "$GUSER:$KC" -X POST "https://gitea.cat-bluegill.ts.net/api/v1/users/$GUSER/tokens" \
     -H "Content-Type: application/json" \
     -d '{"name":"woodpecker-<repo>-ci","scopes":["write:repository"]}'   # response .sha1 == the token
   ```
   Store the `.sha1` value in 1Password **AC-DevOps** (e.g. `gitea - woodpecker <repo> token`).
3. **Add the Woodpecker repo secret `gitea_token`** (restricted to the `tag` event):
   ```bash
   woodpecker-cli repo secret add <org>/<repo> --name gitea_token --value "$GT" --event tag
   woodpecker-cli repo secret ls  <org>/<repo>      # verify
   ```

## The pipeline (`.woodpecker.yml`)

See this repo's [`.woodpecker.yml`](.woodpecker.yml) for the working file. Shape:

- **Build** (`event: [push, pull_request, manual]`) — mirror the local build (`just build-all`):
  one step per artifact, each with `depends_on: []` so they run in parallel.
- **Release** (`event: tag`, `ref: refs/tags/v*`) — two steps:
  - `release-build` (golang image): cross-compile with **literal** `GOOS`/`GOARCH` per target.
  - `release-publish` (alpine + curl/jq): create the Gitea Release and upload the binaries, authed
    with `GITEA_TOKEN` from the `gitea_token` secret. Create-or-reuse the release so re-runs are safe.

## Cutting a release

```bash
# 1. bump the version the binary reports  [repo-specific: here cli/cmd/root.go `var Version`], commit
# 2. tag and push to internal ONLY:
git tag -a vX.Y.Z -m "…"
git push internal vX.Y.Z
```

The tag fires the release pipeline → cross-compiled binaries published to the Gitea Release.

> **Push the tag to `internal` (Gitea), NOT `origin` (GitHub).** A `v*` tag on GitHub would trigger
> the deprecated `.github/workflows/release.yml` goreleaser run. Push `main` to both remotes (no
> trigger); push tags to `internal` only. **[repo-specific to repos mirrored on GitHub]**

## Monitoring & debugging

```bash
woodpecker-cli pipeline ls   <org>/<repo>                  # recent runs
woodpecker-cli pipeline last <org>/<repo>                  # latest
woodpecker-cli pipeline ps   <org>/<repo> <N>              # step states for run N
woodpecker-cli pipeline log show <org>/<repo> <N> <step>   # tail a step's log
woodpecker-cli pipeline start  <org>/<repo> <N>            # re-run run N (same commit)
woodpecker-cli pipeline create <org>/<repo> --branch main  # fresh manual run
```

A pipeline in status **`error` with zero steps** = a config-compile failure; the reason isn't in
the CLI, read it from the API:

```bash
RID=$(curl -s -H "Authorization: Bearer $WOODPECKER_TOKEN" "$WOODPECKER_SERVER/api/repos/lookup/<org>/<repo>" | jq -r .id)
curl -s -H "Authorization: Bearer $WOODPECKER_TOKEN" "$WOODPECKER_SERVER/api/repos/$RID/pipelines/<N>" | jq '{status, errors}'
```

Verify a published release (Gitea API, keychain token — no Touch ID):

```bash
curl -s "https://gitea.cat-bluegill.ts.net/api/v1/repos/<org>/<repo>/releases/tags/vX.Y.Z" \
  -H "Authorization: token $GITEA_TOKEN" | jq '{tag:.tag_name, assets:[.assets[].name]}'
```

## Gotchas (learned during bring-up)

1. **Woodpecker expands `${...}` across the WHOLE config — including comments** — and errors
   (`unable to parse variable name`) on anything it can't parse, compiling the pipeline to **zero
   steps** (status `error`). Keep shell parameter-expansions (`${t%/*}`) and literal `${...}`
   placeholders OUT of `.woodpecker.yml`. Use `$VAR`, `$(...)`, or literals; only `${KNOWN_CI_VAR}`
   forms Woodpecker itself supports are safe.
2. **Cross-compiling in a loop breaks** because of #1 — the loop var's `${t%/*}` gets blanked, so
   `GOOS`/`GOARCH` end up empty. Use one explicit `GOOS=… GOARCH=… go build` line per target.
3. **Steps go parallel under a DAG once any step has `depends_on`.** Give every step an explicit
   `depends_on` (`[]` = independent) so ordering is unambiguous.
4. **Creating a Gitea Release fires a `release` webhook** → a second pipeline runs. Harmless while no
   step matches `event: release` (it no-ops); never add one that could re-trigger a release → loop.
5. **Gitea token creation needs basic auth** (`-u user:pat`), not token auth.
6. **tsidp is login/SSO only** — Gitea's REST API won't accept tsidp bearer tokens for writes. CI →
   Gitea is always a dedicated PAT secret.

## Reusing this in another repo

- Auth is shared: Woodpecker PAT in AC-DevOps, keychain Gitea token, a per-repo `gitea_token` secret.
- Copy `.woodpecker.yml`; swap build/release for that repo's artifacts (binaries → Gitea Releases;
  container images → Harbor, see `direction`'s `.woodpecker/build.yml`).
- A `write:repository` PAT for `ac` works across all orgs, so a token *could* be reused — but per-repo
  least-priv secrets are the house style (independent rotation/revocation).
