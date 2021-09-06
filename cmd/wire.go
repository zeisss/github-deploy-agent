package cmd

import (
	"context"
	"log"

	"github.com/google/go-github/github"
	"github.com/zeisss/github-deploy-agent/agent"

	"golang.org/x/oauth2"
)

func initAgent(ctx context.Context, ownerRepo, env, token, hooksPath string) *agent.Agent {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)

	hooks := &agent.Hooks{
		Path: hooksPath,
	}
	deployments := agent.NewDeploymentAPI(ownerRepo, env, client)
	agent := agent.Agent{
		Log:         log.Default(),
		Deployments: deployments,
		Deployer: &agent.Deployer{
			Hooks:       hooks,
			Deployments: deployments,
			Log:         log.Default(),
		},
	}
	return &agent
}
