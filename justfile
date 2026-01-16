# Mozeidon Development Commands
# Run `just --list` to see all available commands

# Default recipe: list all commands
default:
    @just --list

# ─────────────────────────────────────────────────────────────
# Build Commands
# ─────────────────────────────────────────────────────────────

# Build everything (CLI + Firefox addon + Chrome addon)
build-all:
    make all

# Build CLI only
build-cli:
    make build-cli

# Build Firefox addon only
build-firefox:
    make build-firefox-addon

# Build Chrome addon only
build-chrome:
    make build-chrome-addon

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
    ./cli/mozeidon tabs get

# Get recently closed tabs
tabs-closed:
    ./cli/mozeidon tabs get --closed

# Activate a tab (bring to foreground): just tabs-activate 3289:596
tabs-activate ID:
    ./cli/mozeidon tabs activate {{ID}}

# Get bookmarks
bookmarks:
    ./cli/mozeidon bookmarks

# Get history
history:
    ./cli/mozeidon history

# ─────────────────────────────────────────────────────────────
# Beads (bd) Task Tracking
# ─────────────────────────────────────────────────────────────

# List all open tasks
bd-list:
    bd list

# Show task details: just bd-show mozeidon-abc
bd-show ID:
    bd show {{ID}}

# Close a task with reason: just bd-close mozeidon-abc "Done"
bd-close ID REASON:
    bd close {{ID}} -r "{{REASON}}"

# ─────────────────────────────────────────────────────────────
# Testing Commands
# ─────────────────────────────────────────────────────────────

# Test CLI can connect to Firefox
test-connection:
    ./cli/mozeidon tabs get | head -c 200

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
    cd firefox-addon && rm -f mozeidon-firefox.xpi mozeidon-source.zip && \
    zip -r mozeidon-firefox.xpi manifest.json dist/ icons/ -x "*.DS_Store" && \
    zip -r mozeidon-source.zip src/ package.json package-lock.json webpack.config.js tsconfig.json manifest.json icons/ -x "*.DS_Store" && \
    echo "Created:" && ls -la mozeidon-firefox.xpi mozeidon-source.zip

# Package Chrome extension as .zip
package-chrome:
    cd chrome-addon && zip -r ../mozeidon-chrome.zip manifest.json dist/ assets/

# ─────────────────────────────────────────────────────────────
# Setup Commands
# ─────────────────────────────────────────────────────────────

# Install Firefox native messaging manifest (required for CLI <-> extension communication)
setup-native-messaging:
    mkdir -p ~/Library/Application\ Support/Mozilla/NativeMessagingHosts
    @echo '{"name":"mozeidon_native_app","description":"Mozeidon native messaging host","path":"/opt/homebrew/bin/mozeidon-native-app","type":"stdio","allowed_extensions":["mozeidon-z@a-layer.io","mozeidon@anthropic.github.io","mozeidon-dev@ac.local"]}' > ~/Library/Application\ Support/Mozilla/NativeMessagingHosts/mozeidon_native_app.json
    @echo "Created native messaging manifest. Restart Firefox to apply."
    @cat ~/Library/Application\ Support/Mozilla/NativeMessagingHosts/mozeidon_native_app.json | jq .

# Check if native messaging is configured
check-native-messaging:
    @cat ~/Library/Application\ Support/Mozilla/NativeMessagingHosts/mozeidon_native_app.json 2>/dev/null && echo "\n✓ Native messaging configured" || echo "✗ Native messaging not configured. Run: just setup-native-messaging"

# Full setup: build everything and configure native messaging
setup-all: build-all build-raycast setup-native-messaging
    @echo "\nSetup complete! Next steps:"
    @echo "1. Restart Firefox"
    @echo "2. Install Mozeidon extension in Firefox (about:debugging or AMO)"
    @echo "3. Import Raycast extension from ./raycast/"
