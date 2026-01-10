# Change: Add interactive tab picker with fuzzy search

## Why

Currently, switching to a tab requires knowing its `windowId:tabId` upfront. Users must run `mozeidon tabs get`, visually scan potentially hundreds of tabs, find the ID, then run `mozeidon tabs activate <id>`. This multi-step workflow breaks focus and is slow for power users who expect fzf-style instant filtering.

## What Changes

- **CLI**: Add `mozeidon tabs pick` command (alias: `p`) that launches an interactive TUI
- **Dependencies**: Add Bubbletea (Charm) for TUI, sahilm/fuzzy for fzf-style matching
- **No addon changes**: Uses existing `tabs get` and `tabs activate` commands internally

## Impact

- Affected specs: `tab-picker` (new capability)
- Affected code:
  - `cli/cmd/tabs/pick.go` - new cobra command
  - `cli/cmd/tabs/root.go` - register new command + alias
  - `cli/core/tabs-pick.go` - TUI implementation with bubbletea
  - `go.mod` - new dependencies (bubbletea, lipgloss, fuzzy)
