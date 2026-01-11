# Activate Tab Implementation Guide

This document describes how to implement a feature that brings a specific Firefox tab and its window to the foreground in the Mozeidon ecosystem.

## Current State

Mozeidon has a `switch-tab` command that makes a tab active within Firefox, but it does **not** bring the Firefox window to the foreground over other applications.

```
Current switch-tab behavior:
- Switches to the tab within Firefox ✓
- Brings Firefox window to foreground ✗
```

> **Note:** GitHub issue [#6](https://github.com/egovelox/mozeidon/issues/6) marked this as "impossible to fix" but this is incorrect. The [alfred-firefox](../alfred-firefox/) project proves this works perfectly - including automatic macOS Spaces switching when the target window is on a different Space.

## Goal

Implement an `activate-tab` command that:
1. Makes the specified tab active within its window
2. Brings the Firefox window to the foreground over other applications

## Architecture Overview

```
mozeidon CLI
    │ {command: "activate-tab", args: "tabId"}
    ▼
mozeidon-native-app (no changes needed - transparent proxy)
    │ forwards message via native messaging
    ▼
Mozeidon Firefox addon (NEW: handle activate-tab)
    │ WebExtensions API calls
    ▼
Firefox brings tab and window to foreground
```

## Implementation Details

### 1. Firefox Addon Changes

**Location:** `firefox-addon/`

```
firefox-addon/
├── manifest.json          # Extension manifest
├── src/
│   ├── app.ts             # Entry point, native messaging connection
│   ├── handler.ts         # Command router (switch statement)
│   ├── models/
│   │   └── command.ts     # CommandName enum
│   └── services/
│       └── tabs.ts        # Tab operations
├── package.json           # Dependencies
└── webpack.config.js      # Build config
```

**File:** `firefox-addon/src/models/command.ts`

Add the new command to the enum:

```typescript
export enum CommandName {
  // ... existing commands ...
  ACTIVATE_TAB = "activate-tab",
}
```

**File:** `firefox-addon/src/services/tabs.ts`

Add the activation function:

```typescript
export const activateTab = async (
  port: Browser.Runtime.Port,
  cmd: Command
) => {
  try {
    const tabId = parseInt(cmd.args || "");
    if (isNaN(tabId)) {
      throw new Error("Invalid tab ID");
    }

    // Step 1: Make the tab active within its window
    await browser.tabs.update(tabId, { active: true });

    // Step 2: Get the tab to find its window
    const tab = await browser.tabs.get(tabId);

    // Step 3: Bring the window to the foreground
    await browser.windows.update(tab.windowId, { focused: true });

    port.postMessage(Response.data({ success: true, tabId, windowId: tab.windowId }));
    await delay(10);
    port.postMessage(Response.end());
  } catch (e) {
    handleError(e, port);
  }
};
```

**File:** `firefox-addon/src/handler.ts`

Add the case to the switch statement:

```typescript
import { activateTab } from "./services/tabs";

// In the switch statement:
case CommandName.ACTIVATE_TAB:
  return await activateTab(port, cmd);
```

### 2. CLI Changes

**Location:** `cli/`

```
cli/
├── cmd/
│   └── tabs/              # Tab-related commands
│       └── activate-tab.go  # NEW: activate command
├── core/                  # Implementation logic
│   └── tabs-activate.go   # NEW: activate function
└── browser/
    └── core/
        └── models/
            └── commands.go  # Command struct
```

**File:** `cli/cmd/tabs/activate-tab.go`

```go
package tabs

import (
	"strconv"

	"github.com/egovelox/mozeidon/cli/core"
	"github.com/spf13/cobra"
)

var activateTabCmd = &cobra.Command{
	Use:   "activate [tabId]",
	Short: "Activate a tab and bring its window to foreground",
	Long:  "Makes the specified tab active and brings the Firefox window to the foreground over other applications.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		tabId, err := strconv.Atoi(args[0])
		if err != nil {
			return err
		}
		return core.ActivateTab(tabId)
	},
}

func init() {
	TabsCmd.AddCommand(activateTabCmd)
}
```

**File:** `cli/core/tabs-activate.go`

```go
package core

import (
	"strconv"

	"github.com/egovelox/mozeidon/cli/browser/core/models"
)

func ActivateTab(tabId int) error {
	browser, err := NewBrowser()
	if err != nil {
		return err
	}

	cmd := models.Command{
		Command: "activate-tab",
		Args:    strconv.Itoa(tabId),
	}

	_, err = browser.Send(cmd)
	return err
}
```

### 3. Native App Changes

**None required.** The mozeidon-native-app is a transparent proxy that forwards any `{command, args}` message to the Firefox addon.

## Why Two WebExtensions API Calls?

The implementation requires both API calls for complete tab activation:

| API Call | Purpose |
|----------|---------|
| `browser.tabs.update(tabId, { active: true })` | Makes the tab the active tab within its Firefox window |
| `browser.windows.update(windowId, { focused: true })` | Brings the Firefox window to the foreground over other apps |

Without `windows.update()`, if Firefox is behind other applications, the tab would be activated but invisible to the user.

## Message Flow

```
┌──────────┐     ┌─────────────────┐     ┌─────────────────┐     ┌─────────┐
│ CLI      │     │ Native App      │     │ Firefox Addon   │     │ Firefox │
└────┬─────┘     └───────┬─────────┘     └───────┬─────────┘     └────┬────┘
     │                   │                       │                    │
     │ activate-tab 123  │                       │                    │
     │ (IPC socket)      │                       │                    │
     │──────────────────>│                       │                    │
     │                   │                       │                    │
     │                   │ {cmd:"activate-tab",  │                    │
     │                   │  args:"123"}          │                    │
     │                   │ (native messaging)    │                    │
     │                   │──────────────────────>│                    │
     │                   │                       │                    │
     │                   │                       │ tabs.update(123,   │
     │                   │                       │   {active:true})   │
     │                   │                       │───────────────────>│
     │                   │                       │                    │
     │                   │                       │ tabs.get(123)      │
     │                   │                       │───────────────────>│
     │                   │                       │<───────────────────│
     │                   │                       │                    │
     │                   │                       │ windows.update(    │
     │                   │                       │   wId,{focused:1}) │
     │                   │                       │───────────────────>│
     │                   │                       │                    │
     │                   │                       │<───────────────────│
     │                   │<──────────────────────│                    │
     │<──────────────────│                       │                    │
     │                   │                       │                    │
     │ Tab is now in foreground                  │                    │
```

## macOS-Specific Implementation

On macOS, bringing the correct Firefox window to the foreground requires additional work beyond the WebExtensions API. The `browser.windows.update(windowId, { focused: true })` call focuses the window within Firefox, but doesn't reliably bring it to the front over other applications or switch macOS Spaces.

### The AppleScript Solution

The CLI uses AppleScript to bring the specific window to front:

```applescript
tell application "Firefox"
    activate
    delay 0.1
    set theWindows to every window
    repeat with w in theWindows
        try
            set wName to name of w
            if wName contains "<tab title>" then
                try
                    set index of w to 1
                    return
                end try
            end if
        end try
    end repeat
end tell
```

### Key Implementation Details

1. **Order matters:** Must `activate` Firefox first, then `set index of w to 1`. The reverse order doesn't work reliably.

2. **Delay required:** A 0.1s delay after `activate` ensures the window list is ready.

3. **Nested try blocks:** Some Firefox windows have invalid IDs (id -1). The inner try block catches `set index` errors and continues to the next matching window.

4. **Title matching:** The extension returns the tab title, which becomes the window title. The CLI uses `contains` to match window names.

5. **Title escaping:** Special characters in tab titles are escaped for AppleScript strings.

### Flow

```
Extension                          CLI (macOS)
    │                                  │
    │ browser.tabs.update(active)      │
    │ browser.windows.update(focused)  │
    │ ─────────────────────────────────>│
    │ Response: {tabId, windowId,      │
    │            title: "Tab Title"}   │
    │                                  │
    │                                  │ Sleep 100ms
    │                                  │ osascript: activate + set index
    │                                  │
    │                        Window comes to front
```

## Testing

```bash
# Get list of tabs to find a tab ID
mozeidon tabs get

# Activate a specific tab (brings Firefox to foreground)
mozeidon tabs activate 69:56

# Interactive picker with fuzzy search
mozeidon tabs pick
```

## Repository Locations

All repositories are siblings in the same parent directory:

```
firefox-ai/
├── mozeidon/              # This repo - Firefox addon + CLI (all changes here)
├── mozeidon-native-app/   # Native messaging proxy (no changes needed)
└── alfred-firefox/        # Reference implementation
```

## Related Documentation

- [alfred-firefox TAB_ACTIVATION_ARCHITECTURE.md](../alfred-firefox/TAB_ACTIVATION_ARCHITECTURE.md) - Working implementation that proves this feature is possible
- [GitHub Issue #6](https://github.com/egovelox/mozeidon/issues/6) - Feature request (incorrectly marked as impossible)
- [Firefox WebExtensions tabs.update()](https://developer.mozilla.org/en-US/docs/Mozilla/Add-ons/WebExtensions/API/tabs/update)
- [Firefox WebExtensions windows.update()](https://developer.mozilla.org/en-US/docs/Mozilla/Add-ons/WebExtensions/API/windows/update)
