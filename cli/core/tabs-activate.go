package core

import (
	"os"
	"os/exec"
	"runtime"

	"github.com/egovelox/mozeidon/browser/core/models"
)

func (a *App) TabsActivate(tabId string) {
	returnCode := 0
	done := make(chan bool)

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
			}
		}
		done <- true
	}()

	<-done
	if returnCode != 0 {
		os.Exit(1)
	}

	// On macOS, bring Firefox to foreground via OS-level activation
	// This is needed because browser.windows.update({focused: true}) alone
	// cannot steal focus from another app (macOS security feature)
	if runtime.GOOS == "darwin" {
		cmd := exec.Command("open", "-a", "firefox")
		cmd.Run()
	}
}
