## Context

The mozeidon CLI currently requires users to manually look up tab IDs before switching. This change adds a TUI-based fuzzy picker that combines tab listing and activation into a single interactive command.

**Constraints:**
- Must integrate with existing Go CLI (Cobra-based)
- Must reuse existing IPC for tabs get/activate
- Must work in standard terminals (no GPU/special features)

## Goals / Non-Goals

**Goals:**
- Single command to search and switch tabs
- fzf-style fuzzy filtering
- Minimal latency (<100ms to display list)
- Keyboard-driven (no mouse required)

**Non-Goals:**
- Tab management (close, reorder, group) - out of scope
- Multi-browser support in single view
- Persistent background daemon

## Decisions

### Decision: Use Bubbletea for TUI
- **Why**: Modern, composable, Elm-architecture. Well-maintained by Charm. Already proven in tools like gum, soft-serve.
- **Alternatives considered**:
  - `tview`: Heavier, widget-based, more suited for complex dashboards
  - `go-fuzzyfinder`: Simpler but less customizable
  - `promptui`: Too basic, no fuzzy search built-in

### Decision: Use sahilm/fuzzy for matching
- **Why**: Pure Go, fzf-compatible scoring, simple API
- **Alternatives considered**:
  - Custom implementation: Unnecessary complexity
  - `go-fuzzyfinder` built-in: Couples to their TUI

### Decision: Single file for TUI logic
- **Why**: Keep complexity contained in `cli/core/tabs-pick.go`
- The cobra command in `cli/cmd/tabs/pick.go` will be a thin wrapper

### Decision: Reuse existing TabsGet and TabsActivate
- **Why**: No IPC changes needed, proven code paths
- TUI fetches tabs via internal call to same logic as `tabs get --json`
- Activation uses same `TabsActivate()` function

## Architecture

```
┌─────────────────────────────────────────────┐
│  User runs: mozeidon tabs pick              │
└─────────────────────────────────────────────┘
                    │
                    ▼
┌─────────────────────────────────────────────┐
│  pick.go (Cobra command)                    │
│  - Parse --loop flag                        │
│  - Initialize App                           │
│  - Call TabsPick()                          │
└─────────────────────────────────────────────┘
                    │
                    ▼
┌─────────────────────────────────────────────┐
│  tabs-pick.go (Bubbletea TUI)               │
│  ┌─────────────────────────────────────┐    │
│  │  Model                              │    │
│  │  - tabs []Tab                       │    │
│  │  - filtered []Tab                   │    │
│  │  - cursor int                       │    │
│  │  - query string                     │    │
│  │  - loopMode bool                    │    │
│  └─────────────────────────────────────┘    │
│  ┌─────────────────────────────────────┐    │
│  │  Update()                           │    │
│  │  - Handle keys (j/k/Enter/Esc/R)    │    │
│  │  - Filter on keystroke              │    │
│  └─────────────────────────────────────┘    │
│  ┌─────────────────────────────────────┐    │
│  │  View()                             │    │
│  │  - Render search input              │    │
│  │  - Render tab list with highlights  │    │
│  └─────────────────────────────────────┘    │
└─────────────────────────────────────────────┘
                    │
        ┌───────────┴───────────┐
        ▼                       ▼
┌───────────────┐       ┌───────────────┐
│ TabsGet()     │       │ TabsActivate()│
│ (fetch list)  │       │ (on Enter)    │
└───────────────┘       └───────────────┘
```

## Risks / Trade-offs

| Risk | Mitigation |
|------|------------|
| Bubbletea adds ~5MB to binary | Acceptable for TUI functionality |
| Terminal compatibility | Use lipgloss.AdaptiveColor for light/dark support |
| Large tab counts (1000+) | Lazy rendering, fuzzy filter reduces displayed items |

## Open Questions

None - all clarified during interview.
