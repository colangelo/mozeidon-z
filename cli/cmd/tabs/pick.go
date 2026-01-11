package tabs

import (
	"github.com/spf13/cobra"

	"github.com/egovelox/mozeidon/core"
)

var loopMode bool
var demoMode bool

var PickCmd = &cobra.Command{
	Use:     "pick",
	Aliases: []string{"p"},
	Short:   "Interactive fuzzy tab picker",
	Long: "Launch an interactive TUI to search and switch to browser tabs" +
		"\n\n" +
		"Features:" +
		"\n" +
		"  - Fuzzy search by title and domain" +
		"\n" +
		"  - Tabs sorted by most recently accessed" +
		"\n" +
		"  - Keyboard navigation (arrows/j/k)" +
		"\n" +
		"  - Active tab highlighted with ● marker" +
		"\n\n" +
		"Keyboard shortcuts:" +
		"\n" +
		"  Enter    Select and activate tab" +
		"\n" +
		"  Esc      Cancel and exit" +
		"\n" +
		"  R        Refresh tab list" +
		"\n" +
		"  j/k      Navigate up/down" +
		"\n\n" +
		"Examples:" +
		"\n" +
		"  mozeidon tabs pick" +
		"\n" +
		"  mozeidon tabs p" +
		"\n" +
		"  mozeidon tabs pick --loop" +
		"\n\n",
	Args: cobra.NoArgs,
	Run: func(_ *cobra.Command, args []string) {
		app, err := core.NewApp()
		if err != nil && !demoMode {
			core.PrintError(err.Error())
			return
		}
		if err := app.TabsPick(loopMode, demoMode); err != nil {
			core.PrintError(err.Error())
		}
	},
}

func init() {
	PickCmd.Flags().
		BoolVarP(&loopMode, "loop", "l", false, "Stay open after activating a tab (press Esc to exit)")
	PickCmd.Flags().
		BoolVarP(&demoMode, "demo", "d", false, "Use demo data (for testing without Firefox)")
}
