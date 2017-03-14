package cmd

import (
	"github.com/netlify/gotell/comments"
	"github.com/spf13/cobra"
)

func buildCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "build",
		Short: "build",
		Run: func(cmd *cobra.Command, args []string) {
			execWithConfig(cmd, comments.Build)
		},
	}
}
