package cmd

import (
	"github.com/netlify/gotell/comments"
	"github.com/netlify/gotell/conf"
	"github.com/spf13/cobra"
)

func serveCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "serve",
		Short: "serve",
		Run: func(cmd *cobra.Command, args []string) {
			execWithConfig(cmd, serveComments)
		},
	}
}

func serveComments(config *conf.Configuration) {
	server := comments.NewServer(config)
	server.ListenAndServe()
}
