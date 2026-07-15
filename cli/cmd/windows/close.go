package windows

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/egovelox/mozeidon/cmd/flags"
	"github.com/egovelox/mozeidon/core"
)

var closeYes bool

var CloseWindowCmd = &cobra.Command{
	Use:   "close",
	Short: "Close a given window",
	Long: "Close a given window by id, closing all of its tabs.\n\n" +
		"Required argument: a window id (see `windows get`).\n" +
		"Prompts for confirmation unless -y/--yes is passed (closing a window\n" +
		"takes out all of its tabs at once, so it is guarded by default).\n" +
		"e.g\n" +
		"  mozeidon-z windows close 1\n" +
		"  mozeidon-z windows close 1 -y\n",
	Args: cobra.ExactArgs(1),
	Run: func(_ *cobra.Command, args []string) {
		if !closeYes && !confirmClose(args[0]) {
			fmt.Fprintln(os.Stderr, "Aborted.")
			return
		}
		app, err := core.NewAppWithProfile(flags.ProfileID)
		if err != nil {
			core.PrintError(err.Error())
			return
		}
		app.WindowClose(args[0])
	},
}

// confirmClose asks the user to confirm closing a window. It returns false on
// anything other than an explicit yes — including a non-interactive stdin (EOF),
// so scripts must opt in with -y/--yes rather than being closed silently.
func confirmClose(windowId string) bool {
	fmt.Fprintf(os.Stderr, "Close window %s and all of its tabs? [y/N] ", windowId)
	answer, _ := bufio.NewReader(os.Stdin).ReadString('\n')
	answer = strings.ToLower(strings.TrimSpace(answer))
	return answer == "y" || answer == "yes"
}

func init() {
	CloseWindowCmd.Flags().
		BoolVarP(&closeYes, "yes", "y", false, "skip the confirmation prompt")
}
