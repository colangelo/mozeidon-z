package windows

import (
	"github.com/spf13/cobra"
)

var WindowsCmd = &cobra.Command{
	Use:   "windows",
	Short: "Interact with browser windows",
	Long:  "Retrieve browser windows, focus/create/close them, or change their state.",
}

func init() {
	WindowsCmd.AddCommand(GetWindowsCmd)
	WindowsCmd.AddCommand(FocusWindowCmd)
	WindowsCmd.AddCommand(NewWindowCmd)
	WindowsCmd.AddCommand(CloseWindowCmd)
	WindowsCmd.AddCommand(UpdateWindowCmd)
}
