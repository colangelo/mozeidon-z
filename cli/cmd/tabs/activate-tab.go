package tabs

import (
	"github.com/spf13/cobra"

	"github.com/egovelox/mozeidon/core"
)

var ActivateTabCmd = &cobra.Command{
	Use:   "activate",
	Short: `Activate a given tab and bring window to foreground`,
	Long: "Activate a given tab by id, switching to it and bringing the browser window to the foreground" +
		"\n\n" +
		"Required argument:" +
		"\n" +
		"A string composed of {windowId}:{tabId}" +
		"\n" +
		"e.g" +
		"\n" +
		"mozeidon tabs activate 1:100" +
		"\n\n",
	Args: cobra.ExactArgs(1),
	Run: func(_ *cobra.Command, args []string) {
		app, err := core.NewApp()
		if err != nil {
			core.PrintError(err.Error())
			return
		}
		app.TabsActivate(args[0])
	},
}
