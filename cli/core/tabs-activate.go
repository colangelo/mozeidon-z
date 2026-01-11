package core

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/egovelox/mozeidon/browser/core/models"
)

// activateResponse represents the response from activate-tab command
type activateResponse struct {
	Data struct {
		Success  bool   `json:"success"`
		TabId    int    `json:"tabId"`
		WindowId int    `json:"windowId"`
		Title    string `json:"title"`
	} `json:"data"`
}

func (a *App) TabsActivate(tabId string) {
	returnCode := 0
	done := make(chan bool)
	var windowTitle string

	go func() {
		for result := range a.browser.Send(
			models.Command{
				Command: "activate-tab",
				Args:    tabId,
			},
		) {
			if result.Data != nil {
				if checkForError(result.Data) {
					returnCode = 1
				}
				// Try to extract window title from response
				var resp activateResponse
				if err := json.Unmarshal(result.Data, &resp); err == nil && resp.Data.Title != "" {
					windowTitle = resp.Data.Title
				}
			}
		}
		done <- true
	}()

	<-done
	if returnCode != 0 {
		os.Exit(1)
	}

	// On macOS, bring the specific Firefox window to foreground
	// The browser.windows.update({focused: true}) from the extension focuses the window,
	// but we need AppleScript to bring that specific window to front
	if runtime.GOOS == "darwin" {
		time.Sleep(100 * time.Millisecond)

		if windowTitle != "" {
			// Use AppleScript to find window by title, bring it to front, then activate Firefox
			script := fmt.Sprintf(`
				tell application "Firefox"
					set theWindows to every window
					repeat with w in theWindows
						try
							if name of w contains %q then
								set index of w to 1
								exit repeat
							end if
						end try
					end repeat
					activate
				end tell
			`, escapeAppleScriptString(windowTitle))
			cmd := exec.Command("osascript", "-e", script)
			cmd.Run()
		} else {
			// Fallback: just activate Firefox
			cmd := exec.Command("osascript", "-e", `tell application "Firefox" to activate`)
			cmd.Run()
		}
	}
}

// escapeAppleScriptString escapes special characters for AppleScript strings
func escapeAppleScriptString(s string) string {
	// Truncate to first 50 chars for matching (window titles can be long)
	if len(s) > 50 {
		s = s[:50]
	}
	// Escape backslashes and quotes
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "\"", "\\\"")
	return s
}
