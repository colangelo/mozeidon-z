package windows

import (
	"github.com/spf13/cobra"

	"github.com/egovelox/mozeidon/cmd/flags"
	"github.com/egovelox/mozeidon/core"
)

var loopMode bool
var demoMode bool

var PickWindowCmd = &cobra.Command{
	Use:     "pick",
	Aliases: []string{"p"},
	Short:   "Interactive fuzzy window picker",
	Long: "Launch an interactive TUI to search and focus browser windows" +
		"\n\n" +
		"Features:" +
		"\n" +
		"  - Fuzzy search by active tab title, host and window id" +
		"\n" +
		"  - Last-focused window listed first, marked with ●" +
		"\n" +
		"  - Shows tab count, window id and state per window" +
		"\n" +
		"  - Keyboard navigation (arrows/j/k)" +
		"\n\n" +
		"Keyboard shortcuts:" +
		"\n" +
		"  Enter    Focus the selected window" +
		"\n" +
		"  Esc      Cancel and exit" +
		"\n" +
		"  R        Refresh window list" +
		"\n" +
		"  j/k      Navigate up/down" +
		"\n\n" +
		"Examples:" +
		"\n" +
		"  mozeidon windows pick" +
		"\n" +
		"  mozeidon windows p" +
		"\n" +
		"  mozeidon windows pick --loop" +
		"\n\n",
	Args: cobra.NoArgs,
	Run: func(_ *cobra.Command, _ []string) {
		app, err := core.NewAppWithProfile(flags.ProfileID)
		if err != nil && !demoMode {
			core.PrintError(err.Error())
			return
		}
		if err := app.WindowsPick(loopMode, demoMode); err != nil {
			core.PrintError(err.Error())
		}
	},
}

func init() {
	PickWindowCmd.Flags().
		BoolVarP(&loopMode, "loop", "l", false, "Stay open after focusing a window (press Esc to exit)")
	PickWindowCmd.Flags().
		BoolVarP(&demoMode, "demo", "d", false, "Use demo data (for testing without a browser)")
}
