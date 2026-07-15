package core

import (
	"fmt"
	"os"

	"github.com/egovelox/mozeidon/browser/core/models"
)

// WindowUpdate changes a window's state and/or geometry. Each field is "none" when the
// user did not request a change (args: "<id>:<state>:<top>:<left>:<width>:<height>").
func (a *App) WindowUpdate(windowId int64, state string, top string, left string, width string, height string) {
	if state == "" {
		state = "none"
	}
	args := fmt.Sprintf("%d:%s:%s:%s:%s:%s", windowId, state, top, left, width, height)

	returnCode := 0
	done := make(chan bool)

	go func() {
		for result := range a.browser.Send(
			models.Command{
				Command: "update-window",
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
