package main

import (
	"log"
	"strings"
	"time"

	"github.com/google/go-github/github"
	"github.com/spf13/pflag"
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

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: *token},
	)
	tc := oauth2.NewClient(oauth2.NoContext, ts)

	client := github.NewClient(tc)

	s := strings.Split(*repository, "/")
	owner := s[0]
	repo := s[1]

	deployments := deploymentAPI{
		owner:  owner,
		repo:   repo,
		client: client,
	}
	agent := Agent{
		env:         *env,
		deployments: &deployments,
	}
	if err := agent.run(!*once); err != nil {
		log.Fatalf("Agent failed: %v", err)
	}
}
