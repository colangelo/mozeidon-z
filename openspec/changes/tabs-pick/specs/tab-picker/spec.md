## ADDED Requirements

### Requirement: Interactive Tab Picker Command

The CLI SHALL provide a `mozeidon tabs pick` command that launches an interactive TUI for searching and switching to browser tabs.

#### Scenario: Launch picker and select tab
- **WHEN** user runs `mozeidon tabs pick`
- **THEN** an interactive TUI displays all open tabs sorted by recency
- **AND** user can type to fuzzy-filter the list
- **AND** pressing Enter on a tab activates it and brings Firefox to foreground
- **AND** the picker exits after activation

#### Scenario: Use shorthand alias
- **WHEN** user runs `mozeidon tabs p`
- **THEN** the picker launches identically to `mozeidon tabs pick`

#### Scenario: Cancel without action
- **WHEN** user presses Esc or Ctrl+C in the picker
- **THEN** the picker exits without activating any tab

### Requirement: Fuzzy Search Filtering

The tab picker SHALL provide fzf-style fuzzy search that filters tabs as the user types.

#### Scenario: Filter by partial title match
- **WHEN** user types "git" in the picker
- **THEN** only tabs with titles or domains containing fuzzy match for "git" are shown
- **AND** matched characters are highlighted

#### Scenario: Multiple search terms
- **WHEN** user types "github pr" (space-separated)
- **THEN** only tabs matching BOTH "github" AND "pr" are shown

### Requirement: Visual Tab Indicators

The tab picker SHALL visually distinguish the currently active tab and the selected row.

#### Scenario: Active tab marker
- **WHEN** the picker displays the tab list
- **THEN** the browser's currently active tab is prefixed with ● marker
- **AND** the active tab row uses a distinct accent color

#### Scenario: Selection highlight
- **WHEN** user navigates with arrows or j/k keys
- **THEN** the currently selected row is visually highlighted

### Requirement: Tab List Display Format

Each tab row SHALL display the tab title and domain in a two-column layout.

#### Scenario: Standard tab display
- **WHEN** a tab with title "Pull Request #42" from "github.com" is shown
- **THEN** it displays as: `Pull Request #42                    github.com`

### Requirement: Keyboard Navigation

The picker SHALL support keyboard-only navigation.

#### Scenario: Arrow key navigation
- **WHEN** user presses ↓ or j
- **THEN** selection moves to the next tab

#### Scenario: Vim-style navigation
- **WHEN** user presses k
- **THEN** selection moves to the previous tab

### Requirement: Loop Mode

The picker SHALL support a `--loop` flag for continuous tab switching.

#### Scenario: Stay open after activation
- **WHEN** user runs `mozeidon tabs pick --loop` and presses Enter
- **THEN** the selected tab is activated
- **AND** the picker remains open for another selection
- **AND** user must press Esc to exit

### Requirement: Manual Refresh

The picker SHALL allow manual refresh of the tab list.

#### Scenario: Refresh with R key
- **WHEN** user presses R in the picker
- **THEN** the tab list is re-fetched from the browser
- **AND** the current search filter is preserved

### Requirement: Empty State Handling

The picker SHALL gracefully handle cases with no tabs or connection errors.

#### Scenario: No tabs open
- **WHEN** user launches picker with no browser tabs open
- **THEN** the TUI displays "No tabs found" message
- **AND** user can press Esc to exit

#### Scenario: Connection error
- **WHEN** the browser is not running or IPC fails
- **THEN** the TUI displays "Cannot connect to Firefox" error
- **AND** user can press Esc to exit

### Requirement: Recency Sorting

The tab list SHALL be sorted by most recently accessed first.

#### Scenario: Recently used tabs at top
- **WHEN** the picker displays tabs
- **THEN** tabs are ordered by lastAccessed timestamp descending
- **AND** the most recently visited tab appears first
