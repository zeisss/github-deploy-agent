package main

import (
	"context"
	"log"
	"time"

	"github.com/google/go-github/github"
	"github.com/spf13/pflag"
	"github.com/zeisss/github-deploy-agent/agent"
	"golang.org/x/oauth2"
)

var (
	repository = pflag.String("repository", "", "The repository to deploy for - owner/repo")
	env        = pflag.String("env", "production", "The environment to act on")
	token      = pflag.String("token", "", "API token for github")
	sleepTime  = pflag.Duration("sleep", 60*time.Second, "Sleep time between checks to github api")
	once       = pflag.Bool("once", false, "Check and apploy deployments once, then exit")
)

func main() {
	pflag.Parse()

	ctx := context.Background()

	agent := initAgent(ctx, *repository, *env)
	if err := agent.Run(ctx, !*once, *sleepTime); err != nil {
		log.Fatalf("Agent failed: %v", err)
	}
}

func initAgent(ctx context.Context, ownerRepo, env string) *agent.Agent {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: *token},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)

	deployments := agent.NewDeploymentAPI(ownerRepo, env, client)
	agent := agent.Agent{
		Deployments: deployments,
		Deployer: &agent.Deployer{
			Deployments: deployments,
			Log:         log.Default(),
		},
	}
	return &agent
}
