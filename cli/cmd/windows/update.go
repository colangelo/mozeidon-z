package windows

import (
	"strconv"

	"github.com/spf13/cobra"

	"github.com/egovelox/mozeidon/cmd/flags"
	"github.com/egovelox/mozeidon/core"
)

var updState string
var updTop int64
var updLeft int64
var updWidth int64
var updHeight int64

var UpdateWindowCmd = &cobra.Command{
	Use:   "update",
	Short: "Update a given window's state or geometry",
	Long: "Update a given window by id: change its state (-s) or its geometry.\n\n" +
		"Required argument: a window id (see `windows get`).\n" +
		"State and geometry are mutually exclusive (a maximized/minimized/fullscreen\n" +
		"window cannot also be positioned).\n" +
		"e.g\n" +
		"  mozeidon-z windows update 1 -s maximized\n" +
		"  mozeidon-z windows update 1 --width 1200 --height 800 --top 0 --left 0\n",
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		windowId, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			core.PrintError("Invalid window id: " + args[0])
			return
		}
		if updState != "" && !allowedStates[updState] {
			core.PrintError("Invalid state. Allowed: normal, minimized, maximized, fullscreen")
			return
		}

		// "none" means "leave unchanged"; a geometry value is only sent if the user set it.
		geom := func(name string, v int64) string {
			if cmd.Flag(name).Changed {
				return strconv.FormatInt(v, 10)
			}
			return "none"
		}

		app, err2 := core.NewAppWithProfile(flags.ProfileID)
		if err2 != nil {
			core.PrintError(err2.Error())
			return
		}
		app.WindowUpdate(
			windowId,
			updState,
			geom("top", updTop),
			geom("left", updLeft),
			geom("width", updWidth),
			geom("height", updHeight),
		)
	},
}

func init() {
	UpdateWindowCmd.Flags().
		StringVarP(&updState, "state", "s", "", "window state (normal|minimized|maximized|fullscreen)")
	UpdateWindowCmd.Flags().Int64Var(&updTop, "top", 0, "window top position (px)")
	UpdateWindowCmd.Flags().Int64Var(&updLeft, "left", 0, "window left position (px)")
	UpdateWindowCmd.Flags().Int64Var(&updWidth, "width", 0, "window width (px)")
	UpdateWindowCmd.Flags().Int64Var(&updHeight, "height", 0, "window height (px)")
	UpdateWindowCmd.MarkFlagsOneRequired("state", "top", "left", "width", "height")
	UpdateWindowCmd.MarkFlagsMutuallyExclusive("state", "top")
	UpdateWindowCmd.MarkFlagsMutuallyExclusive("state", "left")
	UpdateWindowCmd.MarkFlagsMutuallyExclusive("state", "width")
	UpdateWindowCmd.MarkFlagsMutuallyExclusive("state", "height")
}
