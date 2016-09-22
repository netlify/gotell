package cmd

import (
	"log"

	"github.com/spf13/cobra"
)

func RootCommand() *cobra.Command {
	rootCmd := cobra.Command{
		Use: "example",
		Run: run,
	}
	rootCmd.AddCommand(BuildCommand(), ServeCommand(), APICommand())
	return &rootCmd
}

func run(cmd *cobra.Command, args []string) {
	log.Printf("goforthandcomment\n\n  build -- builds comments\n  serve -- starts a server\n  api -- start the api server\n")
}
