package agent

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/go-github/github"
)

type Deployer struct {
	Deployments *DeploymentOptions
	Log         Logger
}

func (deployer Deployer) hookContextForDeployment(depl *github.Deployment) Hooks {
	hookEnv := []string{
		fmt.Sprintf("GITHUB_ENV=%s", *depl.Environment),
		fmt.Sprintf("GITHUB_TASK=%s", *depl.Task),
		fmt.Sprintf("GITHUB_DEPLOYMENT_ID=%d", *depl.ID),
		fmt.Sprintf("GITHUB_DEPLOYMENT_URL=%s", *depl.URL),
		fmt.Sprintf("GITHUB_OWNER=%s", deployer.Deployments.owner),
		fmt.Sprintf("GITHUB_REPO=%s", deployer.Deployments.repo),
		fmt.Sprintf("GITHUB_REF=%s", *depl.Ref),
		fmt.Sprintf("GITHUB_SHA=%s", *depl.SHA),
		fmt.Sprintf("GITHUB_CREATOR=%s", *depl.Creator.Login),
		fmt.Sprintf("GITHUB_CREATOR_AVATAR=%s", *depl.Creator.AvatarURL),
	}
	hooks := Hooks{env: hookEnv}
	return hooks
}

func (deployer Deployer) Deploy(ctx context.Context, depl *github.Deployment) error {
	deployer.Log.Printf("Starting deployment=%d...\n", *depl.ID)
	if err := deployer.Deployments.CreateDeploymentStatus(ctx, depl, "pending", "Firing hook"); err != nil {
		return nil
	}
	hooks := deployer.hookContextForDeployment(depl)

	if err := deployer.createDeploymentStatus(ctx, depl, deploymentStateInProgress, "Accepted"); err != nil {
		return err
	}

	if err := hooks.firePreTask(); err != nil {
		deployer.Log.Printf("pre_task failed: %v\n", err)
	}

	if err := hooks.fireCustom(*depl.Task); err != nil {
		var (
			state deploymentState = deploymentStateFailure
			desc  string          = "Hook failed"
		)

		if errors.Is(err, ErrHookNotFound) {
			deployer.Log.Printf("No hook '%s' found.\n", *depl.Task)
			state = deploymentStateError
			desc = "Unknown hook: " + *depl.Task
		}

		if err := hooks.firePostFailure(); err != nil {
			deployer.Log.Printf("post_failure failed: %v\n", err)
		}

		if err := deployer.createDeploymentStatus(ctx, depl, state, desc); err != nil {
			return err
		}
	} else {
		if err := hooks.firePostSuccess(); err != nil {
			deployer.Log.Printf("post_success failed: %v\n", err)
		}

		if err := deployer.createDeploymentStatus(ctx, depl, deploymentStateSuccess, "Finished"); err != nil {
			return err
		}
	}
	return nil
}

func (deployer *Deployer) createDeploymentStatus(ctx context.Context, depl *github.Deployment, state deploymentState, desc string) error {
	deployer.Log.Printf("Setting state=%s descr=%s\n", state, desc)
	return deployer.Deployments.CreateDeploymentStatus(ctx, depl, state, desc)
}
