# Change: Add activate-tab command to bring tab and window to foreground

## Why

The current `switch-tab` command makes a tab active within Firefox, but does **not** bring the Firefox window to the foreground over other applications. This is a significant UX limitation for CLI/automation workflows where users expect to see the tab immediately after switching.

GitHub issue [#6](https://github.com/egovelox/mozeidon/issues/6) marked this as "impossible to fix" - but the [alfred-firefox](../alfred-firefox/) project proves it works perfectly, including automatic macOS Spaces switching.

## What Changes

- **Firefox addon**: Add `ACTIVATE_TAB` command that calls both `browser.tabs.update()` AND `browser.windows.update()` to bring the window to foreground
- **Chrome addon**: Same changes as Firefox (chrome addon source needs extraction from zip first)
- **CLI**: Add `mozeidon tabs activate <windowId:tabId>` command
- **No changes needed** to `mozeidon-native-app` (transparent proxy)

## Impact

- Affected specs: `tab-activation` (new capability)
- Affected code:
  - `firefox-addon/src/models/command.ts` - new enum value
  - `firefox-addon/src/services/tabs.ts` - new `activateTab()` function
  - `firefox-addon/src/handler.ts` - new case
  - `cli/cmd/tabs/activate-tab.go` - new cobra command
  - `cli/cmd/tabs/root.go` - register new command
  - `cli/core/tabs-activate.go` - new core function
  - `chrome-addon/` - mirror Firefox addon changes
