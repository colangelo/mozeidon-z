package core

import (
	"encoding/json"
	"os"
	goTemplates "text/template"

	"github.com/egovelox/mozeidon/browser/core/models"
)

func (a *App) WindowsTemplate(template string) {
	returnCode := 0
	for response := range a.browser.Send(
		models.Command{
			Command: "get-windows",
		},
	) {
		if checkForError(response.Data) {
			returnCode = 1
			continue
		}

		windows := models.Windows{}
		if err := json.Unmarshal(response.Data, &windows); err != nil {
			PrintError("Failed to parse windows data: " + err.Error())
			returnCode = 1
			continue
		}

		t, err := goTemplates.New("windows-template").Parse(template)
		if err != nil {
			PrintError("Invalid template: " + err.Error())
			os.Exit(1)
		}

		if err = t.Execute(os.Stdout, windows); err != nil {
			PrintError("Template execution failed: " + err.Error())
			os.Exit(1)
		}
	}

	if returnCode != 0 {
		os.Exit(1)
	}
}
