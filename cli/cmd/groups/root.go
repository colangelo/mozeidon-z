package groups

import (
	"github.com/spf13/cobra"
)

var GroupsCmd = &cobra.Command{
	Use:     "groups",
	Aliases: []string{"g", "group"},
}

func init() {
	GroupsCmd.AddCommand(GetGroupsCmd)
	GroupsCmd.AddCommand(UpdateGroupCmd)
}
