package cmd

import (
	"context"
	"log"
	"time"

	"github.com/google/go-github/github"
	"github.com/spf13/cobra"
	"github.com/zeisss/github-deploy-agent/agent"
)

var deploymentCmd = cobra.Command{
	Use:   "deployment",
	Short: "Create a new deployment via GitHub API.",
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := CreateDeploymentOptions{
			Log: log.Default(),
		}
		return opts.Run(cmd.Context())
	},
}

var deploymentFlags struct {
	Env       string
	OwnerRepo string
	Branch    string
	NoWait    bool
}

func init() {
	createCmd.AddCommand(&deploymentCmd)

	deploymentCmd.Flags().StringVar(&deploymentFlags.OwnerRepo, "repository", "", "The repository to deploy for - owner/repo")
	deploymentCmd.Flags().StringVar(&deploymentFlags.Env, "env", "testing", "The environment to act on")
	deploymentCmd.Flags().StringVar(&deploymentFlags.Branch, "branch", "master", "Branch, Tag or SHA to create deployment on")
	deploymentCmd.Flags().BoolVar(&deploymentFlags.NoWait, "no-wait", false, "Disable waiting for deployment to reach a final state")
}

type deploymentConfig struct {
	Owner, Repo string
	Client      *github.Client
	Deployments *agent.DeploymentOptions
}

type CreateDeploymentOptions struct {
	Log agent.Logger
}

func (opts CreateDeploymentOptions) Run(ctx context.Context) error {
	config, err := initGithubClient(ctx, deploymentFlags.OwnerRepo, commonConfig.Token)
	if err != nil {
		return err
	}

	req := &github.DeploymentRequest{
		Ref:         &deploymentFlags.Branch,
		Environment: &deploymentFlags.Env,
	}
	deployment, _, err := config.Client.Repositories.CreateDeployment(ctx, config.Owner, config.Repo, req)
	if err != nil {
		return err
	}

	opts.Log.Printf("Deployment created successfully id=%d", *deployment.ID)

	if deploymentFlags.NoWait {
		return nil
	}

	opts.Log.Println("Waiting for deployment to finish")
	t := time.NewTicker(5 * time.Second)

	var states []*github.DeploymentStatus
	defer t.Stop()
	for range t.C {
		states, _, err = config.Client.Repositories.ListDeploymentStatuses(ctx, config.Owner, config.Repo, *deployment.ID, nil)
		if err != nil {
			return err
		}
		if idx := agent.FindState(states, agent.DeploymentStateSuccess, agent.DeploymentStateFailure, agent.DeploymentStateError); idx >= 0 {
			opts.Log.Printf("Deployment finished id=%d state=%s description='%s'", *deployment.ID, *states[idx].State, *states[idx].Description)
			break
		}
	}

	return nil
}
