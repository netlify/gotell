package cmd

import (
	"fmt"
	"log"
	"net/http"

	"github.com/rs/cors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func ServeCommand() *cobra.Command {
	serveCmd := cobra.Command{
		Use:   "serve",
		Short: "serve",
		Run:   ServeComments,
	}
	serveCmd.Flags().IntP("port", "p", 9090, "the port to listen on")

	return &serveCmd
}

func ServeComments(cmd *cobra.Command, args []string) {
	if err := viper.BindPFlags(cmd.Flags()); err != nil {
		log.Fatalf("Failed to read command flags")
	}

	viper.SetEnvPrefix("COMMENT")

	port := viper.GetInt("port")
	handler := cors.Default().Handler(http.FileServer(http.Dir("dist")))
	panic(http.ListenAndServe(fmt.Sprintf(":%v", port), handler))
}
