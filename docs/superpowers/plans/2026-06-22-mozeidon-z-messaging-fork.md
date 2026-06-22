# mozeidon-z-messaging Native-App Fork — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Fork `egovelox/mozeidon-native-app` into `colangelo/mozeidon-z-messaging` — rebrand the binary, add `--version`/`--help`, harden the code, sign + ship via our Homebrew tap, and document how it works — then flip the `mozeidon-z` CLI to depend on it.

**Architecture:** A ~230-line Go native-messaging host that proxies between the browser extension (stdin/stdout) and the CLI (Unix-socket IPC). We rename only the **binary** (`mozeidon-native-app` → `mozeidon-z-messaging`); the native-messaging **host name `"mozeidon"`** and the **IPC socket `mozeidon_native_app`** are frozen, so there is **no AMO change** and no contract break. Released by goreleaser on a `v*` tag.

**Tech Stack:** Go 1.26, `james-barrow/golang-ipc`, `rickypc/native-messaging-host`, `google/uuid`, goreleaser v2, cosign (keyless), GitHub Actions, Homebrew.

## Global Constraints

- **Go version:** build on latest stable **Go 1.26.4**; `go.mod` → `go 1.26`; CI/release `setup-go` uses `stable`.
- **Frozen identifiers (NEVER change):** native-messaging host name **`"mozeidon"`**; IPC socket base name **`mozeidon_native_app`** (and the generated per-instance form `mozeidon_native_app_<pid>_<profileId8>`).
- **Binary name everywhere:** **`mozeidon-z-messaging`**.
- **Module path:** `github.com/colangelo/mozeidon-z-messaging`.
- **License:** MIT. **Platforms:** darwin (universal) + linux (amd64, arm64); **no Windows**.
- **New repo location:** `~/_sync/dev/mozeidon-z-messaging`; GitHub `colangelo/mozeidon-z-messaging` (public); `origin` = colangelo, `upstream` = egovelox.
- **Tap:** `colangelo/homebrew-tap`; secret `HOMEBREW_TAP_TOKEN` (a PAT with write to the tap) must exist on the new repo.
- **stdout is sacred:** it is the native-messaging protocol channel; all logging goes to **stderr** (`log.Printf`).
- **First tag:** `v1.0.0`.
- Commit message trailer on every commit: `Co-Authored-By: Claude Opus 4.8 <noreply@anthropic.com>`.

---

### Task 1: Bootstrap the fork repo (seeded from upstream history)

**Files:** none edited; repo + remotes only.

**Interfaces:**
- Produces: a working Go module at `~/_sync/dev/mozeidon-z-messaging` with `origin`=colangelo (empty GitHub repo) and `upstream`=egovelox, history = egovelox's 3 commits.

- [ ] **Step 1: Create the empty public GitHub repo**

```bash
command gh repo create colangelo/mozeidon-z-messaging --public \
  --description "Mozeidon-Z native-messaging host — browser ⇄ CLI IPC bridge (fork of egovelox/mozeidon-native-app)"
```
Expected: `✓ Created repository colangelo/mozeidon-z-messaging on GitHub`

- [ ] **Step 2: Seed the local checkout from the existing clone, preserving history**

```bash
git clone /Users/ac/_sync/ac-devops/_projects/AI/firefox-ai/mozeidon-native-app \
  /Users/ac/_sync/dev/mozeidon-z-messaging
cd /Users/ac/_sync/dev/mozeidon-z-messaging
git remote set-url origin https://github.com/colangelo/mozeidon-z-messaging.git
git remote add upstream https://github.com/egovelox/mozeidon-native-app.git
git remote -v
git log --oneline -3
```
Expected: `origin` → colangelo, `upstream` → egovelox; log shows `0ea51b4 Supporting multi browser/profiles`, `39660b2`, `100edc9`.

- [ ] **Step 3: Verify the baseline builds and runs as-is**

```bash
cd /Users/ac/_sync/dev/mozeidon-z-messaging
go build -o /tmp/nativeapp-baseline . && echo "BUILD OK"
```
Expected: `BUILD OK` (no behavior change yet; binary name will become correct after Task 2).

- [ ] **Step 4: Push the seeded history to origin**

```bash
git push -u origin HEAD:main
```
Expected: branch `main` pushed to `colangelo/mozeidon-z-messaging`. (No fork commit yet — this is the provenance baseline.)

---

### Task 2: Rename module + binary identity, bump Go, add LICENSE

**Files:**
- Modify: `go.mod` (module path + `go` version)
- Modify: `main.go:16` (import path of `models`)
- Modify: `Makefile`
- Create: `LICENSE`

**Interfaces:**
- Consumes: Task 1 repo.
- Produces: module `github.com/colangelo/mozeidon-z-messaging`; `go build` emits a binary named `mozeidon-z-messaging`. The IPC socket name is **unchanged** (it is generated in `models`, not the module path).

- [ ] **Step 1: Rewrite `go.mod` header**

Change the first two lines of `go.mod` from:
```
module github.com/egovelox/mozeidon-native-app

go 1.21.1
```
to:
```
module github.com/colangelo/mozeidon-z-messaging

go 1.26
```

- [ ] **Step 2: Update the `models` import in `main.go`**

`main.go` line ~16: change
```go
	"github.com/egovelox/mozeidon-native-app/models"
```
to
```go
	"github.com/colangelo/mozeidon-z-messaging/models"
```

- [ ] **Step 3: Replace `Makefile`**

```makefile
build:
	go build -o mozeidon-z-messaging .

test:
	go test ./...

all: build
```

- [ ] **Step 4: Add `LICENSE` (MIT)**

Create `LICENSE` with the standard MIT text, copyright line:
```
MIT License

Copyright (c) 2026 Alfredo Colangelo
Portions Copyright (c) egovelox (original mozeidon-native-app)

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
```

- [ ] **Step 5: Tidy + build to verify the rename**

```bash
go mod tidy
go build -o mozeidon-z-messaging . && ls -la mozeidon-z-messaging && echo "BUILD OK"
```
Expected: `BUILD OK`, binary `mozeidon-z-messaging` present.

- [ ] **Step 6: Commit**

```bash
echo "mozeidon-z-messaging" >> .gitignore   # don't track the built binary
git add go.mod go.sum main.go Makefile LICENSE .gitignore
git commit -m "chore: rename module to mozeidon-z-messaging, Go 1.26, add MIT LICENSE

Co-Authored-By: Claude Opus 4.8 <noreply@anthropic.com>"
```

---

### Task 3: Add `--version` / `--help` (TDD)

**Files:**
- Create: `flags.go`
- Create: `flags_test.go`
- Modify: `main.go` (call `handleFlags` at the top of `main()`; remove any now-duplicate `version`/`IpcIncomingMessage` decls if present)

**Interfaces:**
- Produces: `var version = "dev"` (package main, set via ldflags `-X main.version=…`); `func handleFlags(args []string) (handled bool, output string)`.
- Consumes: nothing.

- [ ] **Step 1: Write the failing test (`flags_test.go`)**

```go
package main

import (
	"strings"
	"testing"
)

func TestHandleFlags_Version(t *testing.T) {
	version = "9.9.9"
	handled, out := handleFlags([]string{"--version"})
	if !handled {
		t.Fatal("expected --version to be handled")
	}
	if out != "mozeidon-z-messaging 9.9.9" {
		t.Fatalf("got %q", out)
	}
}

func TestHandleFlags_VersionShortAlias(t *testing.T) {
	if handled, _ := handleFlags([]string{"-v"}); !handled {
		t.Fatal("expected -v to be handled")
	}
}

func TestHandleFlags_Help(t *testing.T) {
	handled, out := handleFlags([]string{"-h"})
	if !handled {
		t.Fatal("expected -h to be handled")
	}
	if !strings.Contains(out, "native-messaging host") {
		t.Fatalf("help text missing expected phrase: %q", out)
	}
}

func TestHandleFlags_BrowserArgFallsThrough(t *testing.T) {
	// Firefox launches the host with the manifest path + extension id.
	handled, _ := handleFlags([]string{"/path/to/mozeidon.json", "mozeidon-z@a-layer.io"})
	if handled {
		t.Fatal("browser launch args must fall through to the proxy")
	}
}

func TestHandleFlags_NoArgs(t *testing.T) {
	if handled, _ := handleFlags(nil); handled {
		t.Fatal("no args must fall through to the proxy")
	}
}
```

- [ ] **Step 2: Run the test, verify it fails**

```bash
go test ./... -run TestHandleFlags -v
```
Expected: FAIL / build error — `undefined: handleFlags` (and `version`).

- [ ] **Step 3: Implement `flags.go`**

```go
package main

import (
	"fmt"
	"strings"
)

// version is set at build time via -ldflags "-X main.version=…".
var version = "dev"

const helpText = `mozeidon-z-messaging — Mozeidon-Z native-messaging host

A browser native-messaging host: it proxies between the Mozeidon-Z browser
extension (stdin/stdout) and the Mozeidon-Z CLI (Unix-socket IPC). It is
normally launched by the browser, not run directly.

Usage:
  mozeidon-z-messaging [--version] [--help]

Flags:
  -v, --version   print version and exit
  -h, --help      print this help and exit`

// handleFlags inspects process args (os.Args[1:]). If the first arg is a
// recognized flag it returns handled=true plus the text to print. Browser
// launches pass a manifest path / extension id as the first arg, so they
// fall through (handled=false) to the proxy.
func handleFlags(args []string) (handled bool, output string) {
	if len(args) == 0 {
		return false, ""
	}
	switch strings.TrimSpace(args[0]) {
	case "--version", "-v", "version":
		return true, fmt.Sprintf("mozeidon-z-messaging %s", version)
	case "--help", "-h", "help":
		return true, helpText
	default:
		return false, ""
	}
}
```

- [ ] **Step 4: Wire it into `main()` and remove the duplicate `version` if any**

In `main.go`, the body of `main()` must start with:
```go
func main() {
	if handled, out := handleFlags(os.Args[1:]); handled {
		fmt.Println(out)
		os.Exit(0)
	}

	if err := webBrowserProxy(); err != nil {
		log.Printf("Error in mozeidon_native_app: %v", err)
	}
}
```
Ensure `main.go` imports `fmt` and `os` (it already imports `os`; add `fmt` if missing). `version` now lives in `flags.go`, so delete any `version` declaration in `main.go` if one exists.

- [ ] **Step 5: Run the tests, verify they pass**

```bash
go test ./... -run TestHandleFlags -v
```
Expected: PASS (all 5).

- [ ] **Step 6: Manually verify the flag end-to-end**

```bash
go build -o mozeidon-z-messaging .
./mozeidon-z-messaging --version    # → mozeidon-z-messaging dev
./mozeidon-z-messaging --help       # → usage text
```

- [ ] **Step 7: Commit**

```bash
git add flags.go flags_test.go main.go
git commit -m "feat: add --version and --help flags

Co-Authored-By: Claude Opus 4.8 <noreply@anthropic.com>"
```

---

### Task 4: Guard against panic on short profileId (TDD)

**Files:**
- Create: `models/registered-native-app_test.go`
- Modify: `models/registered-native-app.go` (change `GetNativeAppProfile` signature + add guard)
- Modify: `main.go` (handle the new error return)

**Interfaces:**
- Produces: `func GetNativeAppProfile(response *RegistrationInfoResponse) (*NativeAppProfile, error)` — returns an error when `ProfileId` is shorter than 8 chars (previously it panicked on `ProfileId[:8]`).
- Consumes: `RegistrationInfoResponse`, `RegistrationInfo` (unchanged structs in `models/registration-info.go`).

- [ ] **Step 1: Write the failing test (`models/registered-native-app_test.go`)**

```go
package models

import (
	"fmt"
	"os"
	"testing"
)

func respWith(profileId string) *RegistrationInfoResponse {
	return &RegistrationInfoResponse{Data: RegistrationInfo{ProfileId: profileId}}
}

func TestGetNativeAppProfile_Valid(t *testing.T) {
	p, err := GetNativeAppProfile(respWith("12345678-90ab-cdef-1234-567890abcdef"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	wantIpc := fmt.Sprintf("mozeidon_native_app_%d_12345678", os.Getpid())
	if p.IpcName != wantIpc {
		t.Fatalf("IpcName = %q, want %q", p.IpcName, wantIpc)
	}
	wantFile := fmt.Sprintf("%d_12345678.json", os.Getpid())
	if p.FileName != wantFile {
		t.Fatalf("FileName = %q, want %q", p.FileName, wantFile)
	}
}

func TestGetNativeAppProfile_ShortProfileId(t *testing.T) {
	if _, err := GetNativeAppProfile(respWith("abc")); err == nil {
		t.Fatal("expected error for short profileId")
	}
}

func TestGetNativeAppProfile_EmptyProfileId(t *testing.T) {
	if _, err := GetNativeAppProfile(respWith("")); err == nil {
		t.Fatal("expected error for empty profileId")
	}
}
```

- [ ] **Step 2: Run the test, verify it fails**

```bash
go test ./models/ -v
```
Expected: build error — `GetNativeAppProfile` returns 1 value, test expects 2 (and a panic on the short case once that compiles).

- [ ] **Step 3: Add the guard + new signature in `models/registered-native-app.go`**

Change the function signature and prepend the guard; return `nil, err` on failure and `profile, nil` on success:
```go
func GetNativeAppProfile(response *RegistrationInfoResponse) (*NativeAppProfile, error) {
	profileId := response.Data.ProfileId
	if len(profileId) < 8 {
		return nil, fmt.Errorf("invalid profileId %q: must be at least 8 characters", profileId)
	}

	instanceId := uuid.New().String()
	pid := os.Getpid()

	return &NativeAppProfile{
		IpcName:  fmt.Sprintf("mozeidon_native_app_%d_%s", pid, profileId[:8]),
		FileName: fmt.Sprintf("%d_%s.json", pid, profileId[:8]),
		// … remaining fields unchanged (use response.Data.* as before) …
	}, nil
}
```
Keep every other field exactly as before (BrowserName, ProfileRank, etc.). Add `"fmt"` to the imports if not present (it already imports `fmt`, `os`, `path/filepath`, `github.com/google/uuid`).

- [ ] **Step 4: Update the caller in `main.go`**

The current line (~57) `nativeAppProfile = models.GetNativeAppProfile(&registrationData)` becomes:
```go
	nativeAppProfile, err = models.GetNativeAppProfile(&registrationData)
	if err != nil {
		return fmt.Errorf("error building native-app profile: %w", err)
	}
```
(`err` is already declared earlier in `webBrowserProxy`; reuse it. `nativeAppProfile` is already declared as `var nativeAppProfile *models.NativeAppProfile`.)

- [ ] **Step 5: Run tests, verify pass**

```bash
go test ./... -v
```
Expected: PASS (models + flags).

- [ ] **Step 6: Commit**

```bash
git add models/registered-native-app.go models/registered-native-app_test.go main.go
git commit -m "fix: guard against panic when profileId is shorter than 8 chars

Co-Authored-By: Claude Opus 4.8 <noreply@anthropic.com>"
```

---

### Task 5: Robust end-of-stream + non-swallowed codec errors (TDD)

**Files:**
- Create: `proxy_test.go`
- Modify: `main.go` (extract `isEndOfStream`; replace the string-compare terminator; stop ignoring `json.Unmarshal`/`json.Marshal` errors)

**Interfaces:**
- Produces: `func isEndOfStream(response *host.H) bool` (package main) — true iff the response map has `data == "end"`.
- Consumes: `host.H` from `github.com/rickypc/native-messaging-host`.

> **Pre-check:** confirm the map type with `go doc github.com/rickypc/native-messaging-host.H`. It is `map[string]interface{}`; the test below assumes that. If it differs, adjust the literal accordingly.

- [ ] **Step 1: Write the failing test (`proxy_test.go`)**

```go
package main

import (
	"testing"

	host "github.com/rickypc/native-messaging-host"
)

func TestIsEndOfStream(t *testing.T) {
	cases := []struct {
		name string
		h    host.H
		want bool
	}{
		{"end", host.H{"data": "end"}, true},
		{"other", host.H{"data": "more"}, false},
		{"empty", host.H{}, false},
		{"wrong-key", host.H{"foo": "end"}, false},
	}
	for _, c := range cases {
		h := c.h
		if got := isEndOfStream(&h); got != c.want {
			t.Errorf("%s: isEndOfStream = %v, want %v", c.name, got, c.want)
		}
	}
}

func TestIsEndOfStream_Nil(t *testing.T) {
	if isEndOfStream(nil) {
		t.Fatal("nil response must not be end-of-stream")
	}
}
```

- [ ] **Step 2: Run the test, verify it fails**

```bash
go test ./... -run TestIsEndOfStream -v
```
Expected: FAIL — `undefined: isEndOfStream`.

- [ ] **Step 3: Implement `isEndOfStream` and rewire the loop in `main.go`**

Add the function (anywhere in `main.go`, package main):
```go
// isEndOfStream reports whether the browser sent the {"data":"end"} terminator.
// Parses the decoded map instead of byte-comparing marshaled JSON, so key order
// / whitespace can't break streaming.
func isEndOfStream(response *host.H) bool {
	if response == nil {
		return false
	}
	d, ok := (*response)["data"]
	return ok && d == "end"
}
```
Replace the incoming-message unmarshal (currently `json.Unmarshal(message.Data, &incomingMessage)` with ignored error) with:
```go
				incomingMessage := IpcIncomingMessage{}
				if err := json.Unmarshal(message.Data, &incomingMessage); err != nil {
					log.Printf("skipping malformed ipc message: %v", err)
					continue
				}
```
Replace the inner response block (currently `responseMessage, _ := json.Marshal(response)` … `if string(responseMessage) == ` + "`" + `{"data":"end"}` + "`" + ` {`) with:
```go
					responseMessage, err := json.Marshal(response)
					if err != nil {
						return fmt.Errorf("error marshaling browser response: %w", err)
					}
					if err := ipcServer.Write(1, responseMessage); err != nil {
						return fmt.Errorf("error writing to ipc server: %w", err)
					}
					if isEndOfStream(response) {
						break
					}
```
Ensure `main.go` imports `log` and `fmt` (both already used).

- [ ] **Step 4: Run tests, verify pass**

```bash
go test ./... -v
```
Expected: PASS (flags, models, proxy).

- [ ] **Step 5: Commit**

```bash
git add proxy_test.go main.go
git commit -m "fix: parse end-of-stream marker robustly; stop swallowing JSON errors

Co-Authored-By: Claude Opus 4.8 <noreply@anthropic.com>"
```

---

### Task 6: Tighten socket permissions (gated)

**Files:**
- Modify: `main.go` (`ipc.ServerConfig.UnmaskPermissions`)

**Interfaces:** none new. Behavioral change only — verified end-to-end in Task 12.

- [ ] **Step 1: Flip the flag**

In `main.go`, in the `ipc.ServerConfig` literal, change:
```go
		UnmaskPermissions: true, // make the socket writeable for other users (default is false)
```
to:
```go
		UnmaskPermissions: false, // single-user: native-app and CLI run as the same user
```

- [ ] **Step 2: Build (no unit test — IPC perms verified in Task 12)**

```bash
go build -o mozeidon-z-messaging . && echo "BUILD OK"
```
Expected: `BUILD OK`.

- [ ] **Step 3: Commit**

```bash
git add main.go
git commit -m "harden: don't world-write the IPC socket (single-user)

Co-Authored-By: Claude Opus 4.8 <noreply@anthropic.com>"
```

> **Gate:** if Task 12's end-to-end test fails to connect, revert this one commit (`git revert`) and re-test; the rest of the fork stands without it.

---

### Task 7: goreleaser config (rename, tap, cosign, platforms)

**Files:**
- Modify: `.goreleaser.yaml` (full replace)

**Interfaces:**
- Produces: a goreleaser config that builds `mozeidon-z-messaging` for darwin(universal)+linux, ldflags the version into `main.version`, cosign-signs all artifacts, and pushes a formula to `colangelo/homebrew-tap`.

- [ ] **Step 1: Replace `.goreleaser.yaml`**

```yaml
version: 2
project_name: mozeidon-z-messaging

before:
  hooks:
    - go mod tidy

builds:
  - env:
      - CGO_ENABLED=0
    goos:
      - linux
      - darwin
    goarch:
      - amd64
      - arm64
    ldflags:
      - -s -w -X main.version={{ .Version }}

universal_binaries:
  - replace: true

archives:
  - formats: [tar.gz]
    name_template: >-
      {{ .ProjectName }}_
      {{- title .Os }}_
      {{- if eq .Arch "amd64" }}x86_64
      {{- else if eq .Arch "386" }}i386
      {{- else }}{{ .Arch }}{{ end }}

checksum:
  name_template: 'checksums.txt'

changelog:
  sort: asc
  filters:
    exclude:
      - '^docs:'
      - '^test:'
      - '^chore:'

signs:
  - cmd: cosign
    artifacts: all
    output: true
    args:
      - sign-blob
      - '--yes'
      - '--output-signature=${signature}'
      - '--output-certificate=${certificate}'
      - '${artifact}'
    signature: '${artifact}.sig'
    certificate: '${artifact}.pem'

release:
  prerelease: auto

brews:
  - name: mozeidon-z-messaging
    homepage: 'https://github.com/colangelo/mozeidon-z-messaging'
    description: 'Mozeidon-Z native-messaging host — browser ⇄ CLI IPC bridge'
    license: 'MIT'
    repository:
      owner: colangelo
      name: homebrew-tap
      token: '{{ .Env.HOMEBREW_TAP_TOKEN }}'
    commit_author:
      name: github-actions[bot]
      email: github-actions[bot]@users.noreply.github.com
    install: |
      bin.install "mozeidon-z-messaging"
    test: |
      assert_match "mozeidon-z-messaging", shell_output("#{bin}/mozeidon-z-messaging --version")
```

- [ ] **Step 2: Validate the config**

```bash
go install github.com/goreleaser/goreleaser/v2@latest   # if not present
goreleaser check
```
Expected: `1 configuration file(s) validated` / no errors. Fix any schema complaints inline (v2 key names).

- [ ] **Step 3: Local snapshot build (no publish, no sign) to prove builds + brew templating**

```bash
goreleaser release --snapshot --clean --skip=publish,sign
ls dist/ | grep mozeidon-z-messaging
```
Expected: darwin universal + linux amd64/arm64 archives in `dist/`.

- [ ] **Step 4: Commit**

```bash
git add .goreleaser.yaml
git commit -m "ci: retarget goreleaser to mozeidon-z-messaging + our tap + cosign

Co-Authored-By: Claude Opus 4.8 <noreply@anthropic.com>"
```

---

### Task 8: Release + CI workflows

**Files:**
- Modify: `.github/workflows/release.yml` (full replace)
- Create: `.github/workflows/ci.yml`

**Interfaces:**
- Produces: tag-`v*` → goreleaser release (signed + tap bump); push/PR → vet/build/test.

- [ ] **Step 1: Replace `.github/workflows/release.yml`**

```yaml
name: Release

on:
  push:
    tags:
      - 'v[0-9]+.[0-9]+.[0-9]+'
  workflow_dispatch:

permissions:
  contents: write
  id-token: write   # cosign keyless

jobs:
  goreleaser:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0
      - uses: actions/setup-go@v5
        with:
          go-version: 'stable'
      - uses: sigstore/cosign-installer@v3
      - uses: goreleaser/goreleaser-action@v6
        with:
          distribution: goreleaser
          version: '~> v2'
          args: release --clean
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
          HOMEBREW_TAP_TOKEN: ${{ secrets.HOMEBREW_TAP_TOKEN }}
```

- [ ] **Step 2: Create `.github/workflows/ci.yml`**

```yaml
name: CI

on:
  push:
    branches: [main]
  pull_request:

permissions:
  contents: read

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: 'stable'
      - run: go vet ./...
      - run: go build ./...
      - run: go test ./...
```

- [ ] **Step 3: Commit**

```bash
git add .github/workflows/release.yml .github/workflows/ci.yml
git commit -m "ci: GitHub Actions release (goreleaser+cosign) and PR build check

Co-Authored-By: Claude Opus 4.8 <noreply@anthropic.com>"
```

---

### Task 9: Documentation (README, CLAUDE.md, ARCHITECTURE.md)

**Files:**
- Modify: `README.md` (full replace)
- Create: `CLAUDE.md`
- Create: `ARCHITECTURE.md`

**Interfaces:** none (docs).

- [ ] **Step 1: Replace `README.md`**

````markdown
# mozeidon-z-messaging

The native-messaging host for the **Mozeidon-Z** stack — a tiny Go proxy that lets the
[Mozeidon-Z browser extension](https://addons.mozilla.org/firefox/addon/mozeidon-z/) exchange
commands and responses with the [Mozeidon-Z CLI](https://github.com/colangelo/mozeidon-z) over a
local IPC socket.

> Hard fork of [`egovelox/mozeidon-native-app`](https://github.com/egovelox/mozeidon-native-app).
> The **binary** is renamed `mozeidon-z-messaging`; the native-messaging **host name** (`mozeidon`)
> and the **IPC socket** (`mozeidon_native_app`) are unchanged, so it is a drop-in for the
> Mozeidon-Z extension with no AMO change. See [ARCHITECTURE.md](ARCHITECTURE.md) for how it works.

## Install (macOS / Linux)

```bash
brew install colangelo/tap/mozeidon-z-messaging
```

This is also pulled automatically as a dependency of `brew install colangelo/tap/mozeidon-z`.

## Configure native messaging (Firefox, macOS)

Create `~/Library/Application Support/Mozilla/NativeMessagingHosts/mozeidon.json`:

```json
{
  "name": "mozeidon",
  "description": "Mozeidon native messaging host",
  "path": "/opt/homebrew/bin/mozeidon-z-messaging",
  "type": "stdio",
  "allowed_extensions": ["mozeidon-z@a-layer.io"]
}
```

(`just setup-native-messaging` in the `mozeidon-z` repo writes this for you.) Restart Firefox.

## Usage

It is launched by the browser, not run directly. For diagnostics:

```bash
mozeidon-z-messaging --version
mozeidon-z-messaging --help
```

## Build from source

```bash
go build -o mozeidon-z-messaging .   # needs Go 1.26+
```

## Releases

A `v*` git tag triggers GitHub Actions → goreleaser builds darwin (universal) + linux (amd64/arm64),
cosign-signs the artifacts (keyless), publishes a GitHub Release, and bumps the formula in
`colangelo/homebrew-tap`.

## License

MIT. Originally based on `egovelox/mozeidon-native-app`.
````

- [ ] **Step 2: Create `CLAUDE.md`**

````markdown
# CLAUDE.md

Guidance for Claude Code working in this repo.

## What this is

`mozeidon-z-messaging` — the native-messaging host (browser ⇄ CLI IPC bridge) for the Mozeidon-Z
stack. ~230 lines of Go. Hard fork of `egovelox/mozeidon-native-app` (remote `upstream`).

## Frozen identifiers — DO NOT CHANGE

- Native-messaging **host name** `"mozeidon"` — the shipped AMO extension calls
  `connectNative("mozeidon")`. Changing it forces an AMO re-submit.
- IPC **socket** base name `mozeidon_native_app` (generated form
  `mozeidon_native_app_<pid>_<profileId8>`) — contract with the `mozeidon-z` CLI.

Only the **binary filename** (`mozeidon-z-messaging`) is ours to rename.

## Commands

```bash
go build -o mozeidon-z-messaging .   # build
go test ./...                        # test
goreleaser check                     # validate release config
goreleaser release --snapshot --clean --skip=publish,sign   # dry-run release
```

## Release

Bump nothing in code (version comes from the git tag via ldflags). Tag and push:

```bash
git tag -a v1.0.0 -m "mozeidon-z-messaging 1.0.0"
git push origin v1.0.0
```

Needs the `HOMEBREW_TAP_TOKEN` repo secret (PAT with write to `colangelo/homebrew-tap`).

## Architecture

See [ARCHITECTURE.md](ARCHITECTURE.md). stdout is the native-messaging channel — **never** log to
it; use stderr (`log.Printf`).
````

- [ ] **Step 3: Create `ARCHITECTURE.md`**

````markdown
# Architecture — mozeidon-z-messaging

```
Mozeidon-Z browser extension
        ▲  │   native messaging: 4-byte LE length-prefixed JSON over the host's stdin/stdout
        │  ▼
  mozeidon-z-messaging   (this binary — launched by the browser)
        ▲  │   IPC: Unix socket  mozeidon_native_app_<pid>_<profileId8>
        │  ▼
  Mozeidon-Z CLI / Raycast
```

## Two protocols, one proxy

The host bridges two channels:

1. **Browser ↔ host — native messaging.** The browser starts this binary and speaks the
   [native-messaging protocol](https://developer.chrome.com/docs/extensions/develop/concepts/native-messaging):
   each message is a 4-byte little-endian length prefix followed by that many bytes of JSON, over
   the host's **stdin/stdout**. Because stdout *is* the protocol, the host must never print anything
   else there — all logging goes to **stderr** (`log.Printf`).
2. **Host ↔ CLI — IPC.** The host runs a `james-barrow/golang-ipc` server on a Unix socket. The CLI
   connects, sends an `{command, args}` message, and reads streamed responses.

`webBrowserProxy()` is the loop: read an IPC message → forward to the browser
(`PostMessage(os.Stdout, …)`) → read browser responses (`OnMessage(os.Stdin, …)`) and relay each
back over IPC until the `{"data":"end"}` terminator (`isEndOfStream`).

## Registration & multi-profile

On startup the browser sends a first **registration** message (`models.RegistrationInfo`:
browser name/engine/version, `profileId`, rank, name, aliases, user agent, timestamp). The host:

1. Builds a `NativeAppProfile` (`models.GetNativeAppProfile`) with a **per-instance** socket name
   `mozeidon_native_app_<pid>_<profileId8>` and filename `<pid>_<profileId8>.json`.
   (Guarded: a `profileId` shorter than 8 chars is rejected, not panicked on.)
2. Writes that profile JSON into `$UserConfigDir/mozeidon_profiles/`.
3. Starts the IPC server on the per-instance socket.

This lets several browsers/profiles run concurrently, each with its own host instance + socket.

## The 3-way contract

The registration/profile schema is shared across three components — change one, change all three:

| Leg | Component | File |
|---|---|---|
| sends registration | extension | `firefox-addon/src/services/registration.ts` |
| writes profile + socket name | **this host** | `models/*.go` |
| reads profiles, dials the socket | CLI | `mozeidon-z` `cli/profiles/profiles.go`, `cli/core/app.go` |

The CLI also keeps a legacy fallback to the fixed socket `mozeidon_native_app`.

## Lifecycle

On `SIGTERM`/`SIGINT` (the browser closing the host) or any error exit, the host removes its profile
file (`signal.Notify` + `defer os.Remove`), so stale profiles don't accumulate. (Signal-based
unregister does not fire on Windows — which is why we don't ship Windows builds.)

## Security notes

- The IPC socket is created with default (owner-only) permissions (`UnmaskPermissions: false`);
  the host and CLI run as the same user.
- `golang-ipc`'s "encryption" is a homegrown handshake, not audited crypto. The trust model is
  localhost / single-user.
````

- [ ] **Step 4: Commit**

```bash
git add README.md CLAUDE.md ARCHITECTURE.md
git commit -m "docs: README + CLAUDE + ARCHITECTURE (how the bridge works)

Co-Authored-By: Claude Opus 4.8 <noreply@anthropic.com>"
```

---

### Task 10: First release — push, secret, tag, verify

**Files:** none (operational).

**Interfaces:**
- Produces: GitHub Release `v1.0.0` (signed) + formula in `colangelo/homebrew-tap`.

- [ ] **Step 1: Push all fork commits**

```bash
cd /Users/ac/_sync/dev/mozeidon-z-messaging
git push origin main
```

- [ ] **Step 2: Add the tap token secret (value from 1Password — ask the user to run if `op` prompts)**

```bash
command gh secret set HOMEBREW_TAP_TOKEN --repo colangelo/mozeidon-z-messaging \
  --body "$(op read 'op://Private/2mzqpwyhyrgdjfu4g3snrfddjq/rmeijxx743qnyz73kz5gdd4jb4/75psbwbwev74ssiuvkwrkmmqgq')"
command gh secret list --repo colangelo/mozeidon-z-messaging
```
Expected: `HOMEBREW_TAP_TOKEN` listed. (Same PAT used for the `mozeidon-z` CLI tap bump.)

- [ ] **Step 3: Tag and push the release**

```bash
git tag -a v1.0.0 -m "mozeidon-z-messaging 1.0.0"
git push origin v1.0.0
```

- [ ] **Step 4: Watch the run**

```bash
sleep 20
command gh run list --repo colangelo/mozeidon-z-messaging --limit 3
command gh run watch --repo colangelo/mozeidon-z-messaging "$(command gh run list --repo colangelo/mozeidon-z-messaging --json databaseId --jq '.[0].databaseId')" || true
```
Expected: `Release` run succeeds (build → sign → release → brew).

- [ ] **Step 5: Verify the Release + tap formula**

```bash
command gh release view v1.0.0 --repo colangelo/mozeidon-z-messaging \
  --json tagName,assets --jq '{tag:.tagName,assets:[.assets[].name]}'
command gh api repos/colangelo/homebrew-tap/contents/Formula/mozeidon-z-messaging.rb \
  --jq '.content' | base64 -d | grep -E 'version|url|sha256' | head
```
Expected: signed assets present; tap formula has version `1.0.0` and real sha256s.

> **If the brew step fails with exit 128 / auth** → `HOMEBREW_TAP_TOKEN` missing/expired; set it and `gh run rerun <id> --failed --repo colangelo/mozeidon-z-messaging` (same fix pattern as the CLI).

---

### Task 11: Flip the `mozeidon-z` hub to the new bridge

**Files (in `/Users/ac/_sync/ac-devops/_projects/AI/firefox-ai/mozeidon`):**
- Modify: `justfile:159` (manifest `"path"`)
- Modify: `.github/workflows/release.yml` (the `FORMULA_B64` `depends_on`)
- Modify: `README.md`, `WORKSTATION_SETUP.md`, `CLAUDE.md`, `CI_RELEASE_RUNBOOK.md`, `ACTIVATE_TAB_IMPLEMENTATION.md`

**Interfaces:**
- Produces: CLI formula depends on `colangelo/mozeidon-z-messaging`; manifest points at the new binary.

- [ ] **Step 1: Update the native-messaging manifest path in `justfile`**

Line 159: change the manifest `"path"` value from `/opt/homebrew/bin/mozeidon-native-app` to
`/opt/homebrew/bin/mozeidon-z-messaging` (leave `"name":"mozeidon"` and the allowed_extensions
untouched).

- [ ] **Step 2: Re-encode the CLI formula's `depends_on`**

The `mozeidon-z` release workflow embeds the formula as base64 in `FORMULA_B64`
(`.github/workflows/release.yml`). Decode it, change the line
`depends_on "egovelox/mozeidon/mozeidon-native-app"` →
`depends_on "colangelo/mozeidon-z-messaging"`, re-encode, and replace the `FORMULA_B64:` value:

```bash
cd /Users/ac/_sync/ac-devops/_projects/AI/firefox-ai/mozeidon
# decode → edit → re-encode (verify the diff is exactly the depends_on line):
python3 - <<'PY'
import base64, re, pathlib
wf = pathlib.Path(".github/workflows/release.yml")
text = wf.read_text()
m = re.search(r'FORMULA_B64:\s*([A-Za-z0-9+/=]+)', text)
formula = base64.b64decode(m.group(1)).decode()
formula = formula.replace(
    'depends_on "egovelox/mozeidon/mozeidon-native-app"',
    'depends_on "colangelo/mozeidon-z-messaging"')
new_b64 = base64.b64encode(formula.encode()).decode()
wf.write_text(text[:m.start(1)] + new_b64 + text[m.end(1):])
print("OK; new depends_on:", 'depends_on "colangelo/mozeidon-z-messaging"' in formula)
PY
```
Expected: `OK; new depends_on: True`.

- [ ] **Step 3: Update docs (replace where it means *our* binary)**

In `README.md`, `WORKSTATION_SETUP.md`, `CLAUDE.md`, `CI_RELEASE_RUNBOOK.md`,
`ACTIVATE_TAB_IMPLEMENTATION.md`: replace `mozeidon-native-app` with `mozeidon-z-messaging` where it
refers to the bridge **we** ship (the Homebrew dep, the manifest path, the install line). **Keep**
`mozeidon-native-app` in any "relationship to upstream" / egovelox-history prose. Add a `2026-06-22`
audit-log entry to `WORKSTATION_SETUP.md` noting the native-app fork + rename. Update the
component-inventory native-app row to source-of-truth `colangelo/mozeidon-z-messaging`.

- [ ] **Step 4: Commit (in the hub repo)**

```bash
git add justfile .github/workflows/release.yml README.md WORKSTATION_SETUP.md CLAUDE.md CI_RELEASE_RUNBOOK.md ACTIVATE_TAB_IMPLEMENTATION.md
git commit -m "feat: depend on colangelo/mozeidon-z-messaging bridge (forked native-app)

Co-Authored-By: Claude Opus 4.8 <noreply@anthropic.com>"
```

---

### Task 12: End-to-end verification (the real gate)

**Files:** none (operational). Requires Firefox open with the Mozeidon-Z extension.

- [ ] **Step 1: Install the new bridge**

```bash
brew untap colangelo/tap 2>/dev/null; brew tap colangelo/tap
brew install colangelo/tap/mozeidon-z-messaging
which mozeidon-z-messaging          # → /opt/homebrew/bin/mozeidon-z-messaging
mozeidon-z-messaging --version      # → mozeidon-z-messaging 1.0.0
```

- [ ] **Step 2: Point native messaging at it + restart Firefox**

```bash
cd /Users/ac/_sync/ac-devops/_projects/AI/firefox-ai/mozeidon
just setup-native-messaging
command gh   # (no-op) ; then fully quit & reopen Firefox
```

- [ ] **Step 3: Verify the full pipe works (this exercises the renamed binary + socket-perm change)**

```bash
mozeidon-z tabs get | head        # streams JSON of open tabs
pgrep -fl mozeidon-z-messaging    # the bridge process, spawned by Firefox
```
Expected: tabs JSON streams; `mozeidon-z-messaging` process visible. **If `tabs get` cannot connect, revert Task 6** (`git revert` the UnmaskPermissions commit in the fork, re-release a patch, re-install) and re-test.

- [ ] **Step 4: Remove the old egovelox bridge (optional cleanup)**

```bash
brew uninstall egovelox/mozeidon/mozeidon-native-app 2>/dev/null || true
```

- [ ] **Step 5: Final report** — summarize: fork repo + release URL, tap formula, hub `depends_on` flipped, e2e pass/fail, and whether Task 6 stuck or was reverted.

---

## Self-Review

**Spec coverage:** repo bootstrap (T1) ✓; rename/Go/LICENSE (T2) ✓; `--version`/`--help` (T3) ✓;
profileId guard (T4) ✓; end-of-stream + codec errors (T5) ✓; `UnmaskPermissions` (T6) ✓; goreleaser
rename/tap/cosign/platforms/drop-`pro dev` (T7) ✓; release + CI workflows (T8) ✓; README/CLAUDE/
ARCHITECTURE (T9) ✓; first release + secret + verify (T10) ✓; hub `depends_on` + manifest + docs
(T11) ✓; e2e (T12) ✓. Scope cuts (Windows build, stale-file sweep, socket rename) intentionally
absent.

**Placeholder scan:** the `models/registered-native-app.go` change in T4 Step 3 uses "… remaining
fields unchanged …" — this is an instruction to preserve existing struct fields verbatim, not a
code placeholder (the changed lines are shown in full). No TBD/TODO steps.

**Type consistency:** `handleFlags(args []string) (bool, string)`, `version string`,
`GetNativeAppProfile(*RegistrationInfoResponse) (*NativeAppProfile, error)`, `isEndOfStream(*host.H)
bool` — names/signatures consistent across tasks and the docs. Frozen socket name
`mozeidon_native_app_%d_%s` identical in T4 code, test, and ARCHITECTURE.md.
