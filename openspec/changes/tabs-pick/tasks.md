## 1. Dependencies

- [ ] 1.1 Add bubbletea dependency: `go get github.com/charmbracelet/bubbletea`
- [ ] 1.2 Add lipgloss dependency: `go get github.com/charmbracelet/lipgloss`
- [ ] 1.3 Add fuzzy dependency: `go get github.com/sahilm/fuzzy`

## 2. Core TUI Implementation

- [ ] 2.1 Create `cli/core/tabs-pick.go` with bubbletea Model struct
- [ ] 2.2 Implement `Init()` - fetch tabs via existing `TabsGet` logic
- [ ] 2.3 Implement `Update()` - handle keyboard input (j/k/arrows/Enter/Esc/R)
- [ ] 2.4 Implement `View()` - render search input + tab list with styling
- [ ] 2.5 Implement fuzzy filtering using sahilm/fuzzy on title+domain
- [ ] 2.6 Add active tab indicator (● marker + accent color)
- [ ] 2.7 Add match highlighting for fuzzy search results

## 3. Cobra Command

- [ ] 3.1 Create `cli/cmd/tabs/pick.go` with PickCmd definition
- [ ] 3.2 Add `--loop` flag for persistent mode
- [ ] 3.3 Register `PickCmd` in `cli/cmd/tabs/root.go`
- [ ] 3.4 Add `p` as alias for `pick`

## 4. Integration

- [ ] 4.1 Wire tab activation on Enter using existing `TabsActivate()`
- [ ] 4.2 Implement loop mode - refresh and stay open after activation
- [ ] 4.3 Implement R key to manually refresh tab list
- [ ] 4.4 Handle empty state (no tabs) with user-friendly message
- [ ] 4.5 Handle IPC errors with clear error display

## 5. Build & Test

- [ ] 5.1 Build CLI: `make build-cli`
- [ ] 5.2 Test basic flow: `./cli/mozeidon tabs pick` shows list, Enter activates
- [ ] 5.3 Test alias: `./cli/mozeidon tabs p` works
- [ ] 5.4 Test fuzzy search: typing filters the list
- [ ] 5.5 Test loop mode: `--loop` keeps picker open after activation
- [ ] 5.6 Test with 50+ tabs for performance
- [ ] 5.7 Test Esc exits cleanly
