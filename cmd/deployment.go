package cmd

import (
	"context"

	"github.com/google/go-github/github"
	"github.com/spf13/cobra"
)

var deploymentCmd = cobra.Command{
	Use:   "deploy",
	Short: "Create a new deployment via GitHub API.",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runDeployment(cmd.Context())
	},
}

var deploymentFlags struct {
	Env       string
	OwnerRepo string
	Token     string
	Branch    string
}

func init() {
	rootCmd.AddCommand(&deploymentCmd)

	deploymentCmd.Flags().StringVar(&deploymentFlags.OwnerRepo, "repository", "", "The repository to deploy for - owner/repo")
	deploymentCmd.Flags().StringVar(&deploymentFlags.Env, "env", "testing", "The environment to act on")
	deploymentCmd.Flags().StringVar(&deploymentFlags.Token, "token", "", "API token for github")
	deploymentCmd.Flags().StringVar(&deploymentFlags.Branch, "branch", "master", "Branch, Tag or SHA to create deployment on")
}

type deploymentConfig struct {
	Owner, Repo string
	Client      *github.Client
}

func runDeployment(ctx context.Context) error {
	config, err := initGithubClient(ctx, deploymentFlags.OwnerRepo, deploymentFlags.Token)
	if err != nil {
		return err
	}

	req := &github.DeploymentRequest{
		Ref:         &deploymentFlags.Branch,
		Environment: &deploymentFlags.Env,
	}
	_, _, err = config.Client.Repositories.CreateDeployment(ctx, config.Owner, config.Repo, req)
	return err
}
