package cmd

import (
	"context"
	"log"

	"github.com/google/go-github/github"
	"github.com/zeisss/github-deploy-agent/agent"

	"golang.org/x/oauth2"
)

func initAgent(ctx context.Context, ownerRepo, env, token string) *agent.Agent {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
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
