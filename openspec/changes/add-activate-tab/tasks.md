## 1. Firefox Addon

- [ ] 1.1 Add `ACTIVATE_TAB = "activate-tab"` to `CommandName` enum in `firefox-addon/src/models/command.ts`
- [ ] 1.2 Add `activateTab()` function to `firefox-addon/src/services/tabs.ts` that calls `browser.tabs.update()` then `browser.windows.update()`
- [ ] 1.3 Add case for `CommandName.ACTIVATE_TAB` in `firefox-addon/src/handler.ts`
- [ ] 1.4 Build Firefox addon: `make build-firefox-addon`
- [ ] 1.5 Test manually: load temporary addon in Firefox, verify command works via native messaging

## 2. CLI

- [ ] 2.1 Add `cli/core/tabs-activate.go` with `TabsActivate(tabId string)` function
- [ ] 2.2 Add `cli/cmd/tabs/activate-tab.go` with cobra command definition
- [ ] 2.3 Register `ActivateTabCmd` in `cli/cmd/tabs/root.go`
- [ ] 2.4 Build CLI: `make build-cli`
- [ ] 2.5 Test end-to-end: `./cli/mozeidon tabs activate 1:123` brings Firefox to foreground

## 3. Chrome Addon (may defer to follow-up)

- [ ] 3.1 Extract chrome addon source from `source-v3.0.0.zip`
- [ ] 3.2 Add `ACTIVATE_TAB` to Chrome addon `CommandName` enum
- [ ] 3.3 Add `activateTab()` to Chrome addon tabs service
- [ ] 3.4 Add handler case in Chrome addon
- [ ] 3.5 Build Chrome addon: `make build-chrome-addon`
- [ ] 3.6 Test Chrome addon manually

## 4. Validation

- [ ] 4.1 Test Firefox: tab switches AND window comes to foreground
- [ ] 4.2 Test Chrome: tab switches AND window comes to foreground
- [ ] 4.3 Test with Firefox on different macOS Space (should auto-switch)
- [ ] 4.4 Test error handling: invalid tab ID returns error message
