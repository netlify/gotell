package cmd

import (
	"fmt"
	"net/http"
	"strings"

	"golang.org/x/oauth2"

	"github.com/Sirupsen/logrus"
	"github.com/google/go-github/github"
	"github.com/netlify/netlify-comments/api"
	"github.com/netlify/netlify-comments/conf"
	"github.com/spf13/cobra"
)

func apiCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "api",
		Short: "api",
		Run: func(cmd *cobra.Command, args []string) {
			execWithConfig(cmd, serveAPI)
		},
	}
}

func serveAPI(config *conf.Configuration) {
	if err := verifyAPISettings(config); err != nil {
		logrus.Fatalf("Error verifying settings: %v", err)
	}

	if err := verifySite(config.API.SiteURL); err != nil {
		logrus.Fatalf("Error verifying site: %v", err)
	}

	githubClient := newGitHubClient(config)
	if err := verifyRepoAndToken(config.API.Repository, githubClient); err != nil {
		logrus.Fatalf("Error verifying repo: %v", err)
	}

	server := api.NewServerWithVersion(config, githubClient, Version)
	server.ListenAndServe()
}

func verifyAPISettings(config *conf.Configuration) error {
	if config.API.SiteURL == "" {
		return fmt.Errorf("API requires a site url")
	}

	if config.API.Repository == "" {
		return fmt.Errorf("API requires a GitHub repository path")
	}

	if config.API.AccessToken == "" {
		return fmt.Errorf("API requires a GitHub access token")
	}

	return nil
}

func verifySite(url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("Expected 200 status code for %v, got %v", url, resp.StatusCode)
	}
	return nil
}

func verifyRepoAndToken(repository string, client *github.Client) error {
	parts := strings.Split(repository, "/")
	if len(parts) != 2 {
		return fmt.Errorf("Repo format must be owner/repo - %v", repository)
	}
	_, _, err := client.Repositories.Get(parts[0], parts[1])
	return err
}

func newGitHubClient(config *conf.Configuration) *github.Client {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: config.API.AccessToken},
	)
	tc := oauth2.NewClient(oauth2.NoContext, ts)
	return github.NewClient(tc)
}
