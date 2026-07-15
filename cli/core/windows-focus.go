package core

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"time"

	"github.com/egovelox/mozeidon/browser/core/models"
)

// focusResponse represents the response from the focus-window command
type focusResponse struct {
	Data struct {
		Success  bool   `json:"success"`
		WindowId int    `json:"windowId"`
		Title    string `json:"title"`
	} `json:"data"`
}

func (a *App) WindowFocus(windowId string) {
	returnCode := 0
	done := make(chan bool)
	var windowTitle string

	go func() {
		for result := range a.browser.Send(
			models.Command{
				Command: "focus-window",
				Args:    windowId,
			},
		) {
			if result.Data != nil {
				if checkForError(result.Data) {
					returnCode = 1
				}
				// Extract the active tab title so we can raise the specific window
				var resp focusResponse
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

	// On macOS, bring the specific Firefox window to the foreground. The extension's
	// browser.windows.update({focused:true}) focuses it within Firefox, but raising the
	// app window above other apps needs AppleScript (same dance as `tabs activate`).
	if runtime.GOOS == "darwin" {
		time.Sleep(100 * time.Millisecond)

		if windowTitle != "" {
			script := fmt.Sprintf(`
				tell application "Firefox"
					activate
					delay 0.1
					set theWindows to every window
					repeat with w in theWindows
						try
							set wName to name of w
							if wName contains %q then
								try
									set index of w to 1
									return
								end try
							end if
						end try
					end repeat
				end tell
			`, escapeAppleScriptString(windowTitle))
			exec.Command("osascript", "-e", script).Run()
		} else {
			exec.Command("osascript", "-e", `tell application "Firefox" to activate`).Run()
		}
	}
}
