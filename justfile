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
