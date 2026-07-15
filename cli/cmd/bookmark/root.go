package bookmark

import (
	"github.com/spf13/cobra"
)

var BookmarkCmd = &cobra.Command{
	Use:     "bookmark",
	Aliases: []string{"bm"},
}

func init() {
	BookmarkCmd.AddCommand(NewBookmarkCmd)
	BookmarkCmd.AddCommand(DeleteBookmarkCmd)
	BookmarkCmd.AddCommand(UpdateBookmarkCmd)
}
