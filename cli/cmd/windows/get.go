package windows

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/egovelox/mozeidon/cmd/flags"
	"github.com/egovelox/mozeidon/core"
)

var template string

var GetWindowsCmd = &cobra.Command{
	Use:   "get",
	Short: "Get all open windows",
	Long: "Get all open browser windows.\n\n" +
		"Each window reports its id (the same id used as the {windowId} half of a\n" +
		"{windowId}:{tabId} tab id), whether it is focused, its tab count and the\n" +
		"active tab's title/url, plus state, type, private mode and geometry.\n\n" +
		"You may customize output with a go-template using -t, e.g.\n" +
		"  mozeidon-z windows get -t '{{range .Items}}{{.Id}} {{.TabCount}} {{.ActiveTabTitle}}\\n{{end}}'\n",
	Args: cobra.NoArgs,
	Run: func(_ *cobra.Command, _ []string) {
		app, err := core.NewAppWithProfile(flags.ProfileID)
		if err != nil {
			core.PrintError(err.Error())
			return
		}
		if len(template) > 0 {
			app.WindowsTemplate(template)
		} else {
			channelWindows := app.WindowsGet()
			windows := <-channelWindows
			windowsAsString, _ := json.Marshal(windows)
			fmt.Println(string(windowsAsString))
			<-channelWindows
		}
	},
}

func init() {
	GetWindowsCmd.Flags().
		StringVarP(&template, "go-template", "t", "", "go-template to customize output")
}
