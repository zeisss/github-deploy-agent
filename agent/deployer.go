package agent

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/google/go-github/github"
)

type Deployer struct {
	Deployments *DeploymentOptions
}

func (agent Deployer) hookContextForDeployment(depl *github.Deployment) Hooks {
	hookEnv := []string{
		fmt.Sprintf("GITHUB_ENV=%s", *depl.Environment),
		fmt.Sprintf("GITHUB_TASK=%s", *depl.Task),
		fmt.Sprintf("GITHUB_DEPLOYMENT_ID=%d", *depl.ID),
		fmt.Sprintf("GITHUB_DEPLOYMENT_URL=%s", *depl.URL),
		fmt.Sprintf("GITHUB_OWNER=%s", agent.Deployments.owner),
		fmt.Sprintf("GITHUB_REPO=%s", agent.Deployments.repo),
		fmt.Sprintf("GITHUB_REF=%s", *depl.Ref),
		fmt.Sprintf("GITHUB_SHA=%s", *depl.SHA),
		fmt.Sprintf("GITHUB_CREATOR=%s", *depl.Creator.Login),
		fmt.Sprintf("GITHUB_CREATOR_AVATAR=%s", *depl.Creator.AvatarURL),
	}
	hooks := Hooks{env: hookEnv}
	return hooks
}

func (agent Deployer) Deploy(ctx context.Context, depl *github.Deployment) error {
	log.Printf("Starting deployment=%d...\n", *depl.ID)
	if err := agent.Deployments.CreateDeploymentStatus(ctx, depl, "pending", "Firing hook"); err != nil {
		return nil
	}
	hooks := agent.hookContextForDeployment(depl)

	if err := hooks.firePreTask(); err != nil {
		log.Printf("pre_task failed: %v\n", err)
	}

	if err := hooks.fireCustom(*depl.Task); err != nil {
		var (
			state string = "failure"
			desc  string = "Hook failed"
		)

		if errors.Is(err, ErrHookNotFound) {
			log.Printf("No hook '%s' found.\n", *depl.Task)
			state = "error"
			desc = "Unknown hook: " + *depl.Task
		}

		if err := hooks.firePostFailure(); err != nil {
			log.Printf("post_failure failed: %v\n", err)
		}

		if err := agent.Deployments.CreateDeploymentStatus(ctx, depl, state, desc); err != nil {
			return nil
		}
	} else {
		if err := hooks.firePostSuccess(); err != nil {
			log.Printf("post_success failed: %v\n", err)
		}

		if err := agent.Deployments.CreateDeploymentStatus(ctx, depl, "success", "Finished"); err != nil {
			return nil
		}
	}
	return nil
}
