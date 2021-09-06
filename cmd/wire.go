package cmd

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/google/go-github/github"
	"github.com/zeisss/github-deploy-agent/agent"

	"golang.org/x/oauth2"
)

func initGithubClient(ctx context.Context, ownerRepo, token string) (*deploymentConfig, error) {
	client := provideGithubClient(ctx, token)
	owner, repo, err := provideOwnerAndRepo(ownerRepo)
	if err != nil {
		return nil, err
	}

	return &deploymentConfig{
		Client: client,
		Owner:  owner,
		Repo:   repo,
	}, nil
}

func initAgent(ctx context.Context, ownerRepo, env, token, hooksPath string) (*agent.Agent, error) {
	client := provideGithubClient(ctx, token)
	owner, repo, err := provideOwnerAndRepo(ownerRepo)
	if err != nil {
		return nil, err
	}

	hooks := &agent.Hooks{
		Path: hooksPath,
	}
	deployments := &agent.DeploymentOptions{
		Owner:  owner,
		Repo:   repo,
		Client: client,
		Env:    env,
	}
	agent := agent.Agent{
		Log:         log.Default(),
		Deployments: deployments,
		Deployer: &agent.Deployer{
			Hooks:       hooks,
			Deployments: deployments,
			Log:         log.Default(),
		},
	}
	return &agent, nil
}

func provideGithubClient(ctx context.Context, token string) *github.Client {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)

	return client
}

func provideOwnerAndRepo(input string) (string, string, error) {
	s := strings.Split(input, "/")
	if len(s) != 2 {
		return "", "", fmt.Errorf("expected <owner>/<repo> format, got '%s'", input)
	}
	owner := s[0]
	repo := s[1]
	return owner, repo, nil
}
