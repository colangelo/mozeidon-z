package windows

import (
	"github.com/spf13/cobra"

	"github.com/egovelox/mozeidon/cmd/flags"
	"github.com/egovelox/mozeidon/core"
)

var FocusWindowCmd = &cobra.Command{
	Use:   "focus",
	Short: "Focus a given window and bring it to the foreground",
	Long: "Focus a given window by id, bringing it to the foreground.\n\n" +
		"Required argument: a window id (see `windows get`).\n" +
		"e.g\n" +
		"  mozeidon-z windows focus 1\n",
	Args: cobra.ExactArgs(1),
	Run: func(_ *cobra.Command, args []string) {
		app, err := core.NewAppWithProfile(flags.ProfileID)
		if err != nil {
			core.PrintError(err.Error())
			return
		}
		app.WindowFocus(args[0])
	},
}
