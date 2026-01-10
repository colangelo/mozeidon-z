## Context

Mozeidon's `switch-tab` command calls only `browser.tabs.update(tabId, { active: true })`, which makes the tab active within Firefox but doesn't bring the Firefox window to the foreground if it's behind other applications.

The alfred-firefox project implements `activateTab()` that solves this by calling two WebExtensions APIs in sequence:

```javascript
// alfred-firefox/extension/alfred.js:320-328
self.activateTab = id => {
  return browser.tabs
    .update(id, { active: true })
    .then(() => {
      return browser.tabs.get(id);
    })
    .then(tab => {
      return browser.windows.update(tab.windowId, { focused: true });
    });
};
```

## Goals / Non-Goals

**Goals:**
- Add new `activate-tab` command that brings tab AND window to foreground
- Maintain backward compatibility - existing `switch-tab` behavior unchanged
- Support both Firefox and Chrome

**Non-Goals:**
- Modifying existing `switch-tab` command (would break existing scripts)
- Removing the `--open` flag from `switch-tab` (keep as legacy option)

## Decisions

### Decision 1: New command vs modifying existing

**Choice:** Create new `activate-tab` command instead of modifying `switch-tab`

**Rationale:**
- Backward compatibility: existing scripts using `switch-tab` continue to work
- Clear semantics: `switch` = change active tab, `activate` = bring to foreground
- The `--open` flag on `switch-tab` is a macOS-only workaround; `activate` works cross-platform via WebExtensions API

### Decision 2: Tab ID format

**Choice:** Accept same `windowId:tabId` format as other commands (for consistency), but also support plain `tabId`

**Rationale:**
- alfred-firefox uses plain `tabId` because `browser.tabs.update()` + `browser.tabs.get()` only need the tab ID
- However, mozeidon convention is `windowId:tabId` for all tab operations
- We can parse both formats for flexibility

### Decision 3: Response format

**Choice:** Return `{ success: true, tabId, windowId }` on success

**Rationale:**
- Confirms which tab/window was activated
- Useful for debugging and scripting

## Risks / Trade-offs

| Risk | Mitigation |
|------|------------|
| Chrome addon source is in zip file | Extract and update, or defer Chrome support to follow-up |
| Window focus may not work on all platforms | WebExtensions API is standardized; tested on macOS |

## Implementation Sequence

1. Firefox addon changes (extension first, so we can test via CLI)
2. CLI changes
3. Chrome addon changes (may be separate PR if source extraction is complex)

## Open Questions

None - the alfred-firefox implementation proves the approach works.
