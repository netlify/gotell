package cmd

import (
	"github.com/Sirupsen/logrus"
	"github.com/netlify/netlify-comments/conf"
	"github.com/spf13/cobra"
)

func RootCommand() *cobra.Command {
	rootCmd := cobra.Command{
		Use: "netlify-comments",
		Run: run,
	}

	rootCmd.AddCommand(buildCommand(), serveCommand(), apiCommand(), &versionCmd)
	rootCmd.PersistentFlags().StringP("config", "c", "", "the config file to use")

	return &rootCmd
}

func run(cmd *cobra.Command, args []string) {
	execWithConfig(cmd, serveAPI)
}

func execWithConfig(cmd *cobra.Command, fn func(config *conf.Configuration)) {
	config, err := conf.LoadConfig(cmd)
	if err != nil {
		logrus.Fatalf("Failed to load configration: %+v", err)
	}

	fn(config)
}
