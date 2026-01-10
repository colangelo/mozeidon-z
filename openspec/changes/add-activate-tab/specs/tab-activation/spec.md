## ADDED Requirements

### Requirement: Tab Activation with Window Focus

The system SHALL provide an `activate-tab` command that makes a specified tab active AND brings its containing browser window to the foreground over other applications.

#### Scenario: Activate tab when browser is in background

- **WHEN** the user runs `mozeidon tabs activate <windowId:tabId>` while the browser window is behind other applications
- **THEN** the specified tab becomes the active tab within its window
- **AND** the browser window is brought to the foreground over all other applications

#### Scenario: Activate tab on different macOS Space

- **WHEN** the user runs `mozeidon tabs activate <windowId:tabId>` and the target window is on a different macOS Space
- **THEN** macOS switches to the Space containing the browser window
- **AND** the specified tab is activated

#### Scenario: Activate tab with plain tab ID

- **WHEN** the user runs `mozeidon tabs activate <tabId>` (without windowId prefix)
- **THEN** the system looks up the window containing the tab
- **AND** activates the tab and brings its window to foreground

#### Scenario: Invalid tab ID

- **WHEN** the user runs `mozeidon tabs activate <invalidId>`
- **THEN** the system returns an error message indicating the tab was not found

### Requirement: Tab Activation Response

The system SHALL return confirmation of successful tab activation.

#### Scenario: Success response

- **WHEN** the activate-tab command succeeds
- **THEN** the response includes `{ success: true, tabId: <id>, windowId: <id> }`

### Requirement: Backward Compatibility

The existing `switch-tab` command SHALL remain unchanged, only making the tab active without bringing the window to foreground.

#### Scenario: switch-tab behavior preserved

- **WHEN** the user runs `mozeidon tabs switch <windowId:tabId>`
- **THEN** the tab becomes active within its window
- **AND** the window focus is NOT changed (browser may remain in background)
