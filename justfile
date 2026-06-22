# Mozeidon Development Commands
# Run `just --list` to see all available commands

# Default recipe: list all commands
default:
    @just --list

# ─────────────────────────────────────────────────────────────
# Build Commands
# ─────────────────────────────────────────────────────────────

# Build everything (CLI + Firefox addon + Chrome addon)
build-all: build-cli build-firefox build-chrome

# Build CLI only (→ cli/mozeidon-z)
build-cli:
    cd cli && go build -o mozeidon-z

# Build and install CLI to ~/.local/bin
install-cli:
    cd cli && go build -o ~/.local/bin/mozeidon-z .

# Build Firefox addon (also copies the webextension-polyfill source map to silence a Firefox warning)
build-firefox:
    cd firefox-addon && npm install && npm run prettier && npm run build
    cd firefox-addon && cp node_modules/webextension-polyfill/dist/browser-polyfill.js.map dist/ 2>/dev/null || true

# Build Chrome addon (src is synced verbatim from firefox-addon, then built)
build-chrome:
    rm -rf chrome-addon/src
    cp -r firefox-addon/src chrome-addon/src
    cd chrome-addon && npm install && npm run build

# Build Raycast extension
build-raycast:
    cd raycast && npm install && npm run build

# ─────────────────────────────────────────────────────────────
# Development Commands
# ─────────────────────────────────────────────────────────────

# Run Raycast extension in dev mode (hot reload)
raycast-dev:
    cd raycast && npm run dev

# Lint Raycast extension
raycast-lint:
    cd raycast && npm run lint

# Format Firefox addon TypeScript
format-firefox:
    cd firefox-addon && npm run prettier

# ─────────────────────────────────────────────────────────────
# CLI Commands
# ─────────────────────────────────────────────────────────────

# Get open tabs
tabs-get:
    ./cli/mozeidon-z tabs get

# Get recently closed tabs
tabs-closed:
    ./cli/mozeidon-z tabs get --closed

# Activate a tab (bring to foreground): just tabs-activate 3289:596
tabs-activate ID:
    ./cli/mozeidon-z tabs activate {{ID}}

# Get bookmarks
bookmarks:
    ./cli/mozeidon-z bookmarks

# Get history
history:
    ./cli/mozeidon-z history

# ─────────────────────────────────────────────────────────────
# Testing Commands
# ─────────────────────────────────────────────────────────────

# Test CLI can connect to Firefox
test-connection:
    ./cli/mozeidon-z tabs get | head -c 200

# Open Firefox debugging page (for loading local extension)
firefox-debug:
    open "about:debugging#/runtime/this-firefox"

# Open Chrome extensions page
chrome-extensions:
    open "chrome://extensions/"

# ─────────────────────────────────────────────────────────────
# Git Commands
# ─────────────────────────────────────────────────────────────

# Show unpushed commits
git-unpushed:
    git log origin/main..HEAD --oneline

# Push to origin
git-push:
    git push

# Show status
git-status:
    git status

# ─────────────────────────────────────────────────────────────
# Extension Packaging
# ─────────────────────────────────────────────────────────────

# Package Firefox extension for AMO upload (.xpi + source.zip)
package-firefox:
    #!/usr/bin/env bash
    cd firefox-addon
    VERSION=$(grep '"version"' manifest.json | sed 's/.*: "\(.*\)".*/\1/')
    XPI_NAME="mozeidon-z-${VERSION}.xpi"
    SRC_NAME="mozeidon-z-${VERSION}-source.zip"
    rm -f mozeidon-z-*.xpi mozeidon-z-*-source.zip
    zip -r "$XPI_NAME" manifest.json dist/ icons/ -x "*.DS_Store"
    zip -r "$SRC_NAME" src/ package.json package-lock.json webpack.config.js tsconfig.json manifest.json icons/ -x "*.DS_Store"
    echo "Created:"
    ls -la "$XPI_NAME" "$SRC_NAME"

# Submit a NEW Firefox version to AMO (listed): build → package → sign & publish.
# All installs auto-update once AMO approves. Bump manifest.json "version" first
# (must exceed the published version). Reads AMO JWT creds from the environment:
#   WEB_EXT_API_KEY (issuer), WEB_EXT_API_SECRET (secret) — never pass on the CLI.
submit-firefox: build-firefox package-firefox
    #!/usr/bin/env bash
    set -euo pipefail
    : "${WEB_EXT_API_KEY:?set WEB_EXT_API_KEY (AMO JWT issuer)}"
    : "${WEB_EXT_API_SECRET:?set WEB_EXT_API_SECRET (AMO JWT secret)}"
    cd firefox-addon
    VERSION=$(grep '"version"' manifest.json | sed 's/.*: "\(.*\)".*/\1/')
    echo "Submitting mozeidon-z ${VERSION} to AMO (listed channel)…"
    rm -rf .amo-pkg && mkdir -p .amo-pkg
    cp -r manifest.json dist icons .amo-pkg/
    # web-ext reads WEB_EXT_API_KEY / WEB_EXT_API_SECRET from the env (keeps them out of argv)
    npx --yes web-ext@latest sign \
      --source-dir=.amo-pkg --channel=listed \
      --upload-source-code="mozeidon-z-${VERSION}-source.zip"
    rm -rf .amo-pkg
    echo "✓ Submitted ${VERSION}. AMO reviews/signs; installs auto-update once approved."

# Package Chrome extension as .zip
package-chrome:
    cd chrome-addon && zip -r ../mozeidon-chrome.zip manifest.json dist/ assets/

# ─────────────────────────────────────────────────────────────
# Setup Commands
# ─────────────────────────────────────────────────────────────

# Install Firefox native messaging manifest (required for CLI <-> extension communication)
setup-native-messaging:
    mkdir -p ~/Library/Application\ Support/Mozilla/NativeMessagingHosts
    @echo '{"name":"mozeidon","description":"Mozeidon native messaging host","path":"/opt/homebrew/bin/mozeidon-z-messaging","type":"stdio","allowed_extensions":["mozeidon-z@a-layer.io","mozeidon@anthropic.github.io","mozeidon-dev@ac.local"]}' > ~/Library/Application\ Support/Mozilla/NativeMessagingHosts/mozeidon.json
    @echo "Created native messaging manifest. Restart Firefox to apply."
    @cat ~/Library/Application\ Support/Mozilla/NativeMessagingHosts/mozeidon.json | jq .

# Check if native messaging is configured
check-native-messaging:
    @cat ~/Library/Application\ Support/Mozilla/NativeMessagingHosts/mozeidon.json 2>/dev/null && echo "\n✓ Native messaging configured" || echo "✗ Native messaging not configured. Run: just setup-native-messaging"

# Full setup: build everything and configure native messaging
setup-all: build-all build-raycast setup-native-messaging
    @echo "\nSetup complete! Next steps:"
    @echo "1. Restart Firefox"
    @echo "2. Install Mozeidon extension in Firefox (about:debugging or AMO)"
    @echo "3. Import Raycast extension from ./raycast/"
