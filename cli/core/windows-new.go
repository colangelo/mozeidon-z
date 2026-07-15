package core

import (
	"fmt"
	"os"
	"strings"

	"github.com/egovelox/mozeidon/browser/core/models"
)

// WindowNew opens a new browser window, optionally with a set of urls, in a private
// (incognito) window, and/or in a given state (normal|minimized|maximized|fullscreen).
func (a *App) WindowNew(urls []string, incognito bool, state string) {
	if state == "" {
		state = "none"
	}
	// flags and urls are separated by "|"; individual urls by "\n" (urls contain ":"/"," but not "\n").
	args := fmt.Sprintf("%t:%s|%s", incognito, state, strings.Join(urls, "\n"))

	returnCode := 0
	done := make(chan bool)

	go func() {
		for result := range a.browser.Send(
			models.Command{
				Command: "new-window",
				Args:    args,
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
}
