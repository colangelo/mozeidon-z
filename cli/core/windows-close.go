package core

import (
	"os"

	"github.com/egovelox/mozeidon/browser/core/models"
)

func (a *App) WindowClose(windowId string) {
	returnCode := 0
	done := make(chan bool)

	go func() {
		for result := range a.browser.Send(
			models.Command{
				Command: "close-window",
				Args:    windowId,
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
