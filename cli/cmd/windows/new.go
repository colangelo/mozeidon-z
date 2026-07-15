package windows

import (
	"github.com/spf13/cobra"

	"github.com/egovelox/mozeidon/cmd/flags"
	"github.com/egovelox/mozeidon/core"
)

var newUrls []string
var private bool
var newState string

// allowed window states for windows.create / windows.update
var allowedStates = map[string]bool{
	"normal":     true,
	"minimized":  true,
	"maximized":  true,
	"fullscreen": true,
}

var NewWindowCmd = &cobra.Command{
	Use:   "new",
	Short: "Open a new window",
	Long: "Open a new browser window.\n\n" +
		"Optionally open one or more urls (repeat -u), open a private window with -p,\n" +
		"and/or set the initial state with -s (normal|minimized|maximized|fullscreen).\n" +
		"e.g\n" +
		"  mozeidon-z windows new -u https://example.com -u https://mozilla.org\n" +
		"  mozeidon-z windows new -p -s maximized\n",
	Args: cobra.NoArgs,
	Run: func(_ *cobra.Command, _ []string) {
		if newState != "" && !allowedStates[newState] {
			core.PrintError("Invalid state. Allowed: normal, minimized, maximized, fullscreen")
			return
		}
		app, err := core.NewAppWithProfile(flags.ProfileID)
		if err != nil {
			core.PrintError(err.Error())
			return
		}
		app.WindowNew(newUrls, private, newState)
	},
}

func init() {
	NewWindowCmd.Flags().
		StringArrayVarP(&newUrls, "url", "u", []string{}, "url to open (repeatable)")
	NewWindowCmd.Flags().
		BoolVarP(&private, "private", "p", false, "open a private (incognito) window")
	NewWindowCmd.Flags().
		StringVarP(&newState, "state", "s", "", "initial window state (normal|minimized|maximized|fullscreen)")
}
