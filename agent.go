package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/giantswarm/retry-go"
	"github.com/google/go-github/github"
)

type deploymentAPI struct {
	owner  string
	repo   string
	client *github.Client
}

func (api deploymentAPI) getDeployments(ctx context.Context, env string) (deployments []*github.Deployment, err error) {
	retry.Do(func() error {
		deployments, _, err = api.client.Repositories.ListDeployments(ctx, api.owner, api.repo, &github.DeploymentsListOptions{
			Environment: env,
		})
		return err
	})
	return deployments, err
}

func (api deploymentAPI) findNewestDeployment(ctx context.Context, env string) (*github.Deployment, error) {
	deployments, err := api.getDeployments(ctx, env)
	if err != nil {
		return nil, err
	}

	var newestDeployment *github.Deployment
	for _, deployment := range deployments {
		if newestDeployment.CreatedAt == nil || deployment.CreatedAt.Time.After(newestDeployment.CreatedAt.Time) {
			newestDeployment = deployment
		}
	}
	if newestDeployment.ID == nil {
		return nil, nil
	}
	return newestDeployment, nil
}

func (api deploymentAPI) hasSuccessStatus(ctx context.Context, depl *github.Deployment) (bool, error) {
	statuses, _, err := api.client.Repositories.ListDeploymentStatuses(ctx, api.owner, api.repo, *depl.ID, &github.ListOptions{})
	if err != nil {
		return false, err
	}

	for _, status := range statuses {
		if status.State != nil && *status.State == "success" {
			return true, nil
		}
	}
	return false, nil
}

// createDeploymentStatus publishes a new status message for the given deployment object.
//
// see https://developer.github.com/v3/repos/deployments/#create-a-deployment-status
// state = pending | success | error | failure
// description = string(140)
func (api deploymentAPI) createDeploymentStatus(ctx context.Context, depl *github.Deployment, state, desc string) error {
	log.Printf("Setting state=%s descr=%s\n", state, desc)
	return retry.Do(func() error {
		_, _, err := api.client.Repositories.CreateDeploymentStatus(ctx, api.owner, api.repo, *depl.ID, &github.DeploymentStatusRequest{
			State:       &state,
			Description: &desc,
		})
		return err
	})
}

// Agent applies deployments
type Agent struct {
	env         string
	deployments *deploymentAPI
}

func (agent Agent) run(ctx context.Context, loop bool) error {
	deployment, err := agent.deployments.findNewestDeployment(ctx, agent.env)
	if err != nil {
		return err
	}

	var lastID int64
	if deployment != nil {
		// If the latest deployment has no success message, deploy it immediately
		if success, err := agent.deployments.hasSuccessStatus(ctx, deployment); err != nil {
			return err
		} else if !success {
			log.Printf("Found new deployment=%d\n", *deployment.ID)
			if err := agent.deploy(ctx, deployment); err != nil {
				return err
			}
		} else {
			log.Printf("Latest deployment %d has success message. Using as baseline.", *deployment.ID)
		}
		lastID = *deployment.ID
	} else {
		log.Println("No deployment in repository found.")
	}

	if loop {
		for {
			if lastID, err = agent.checkRepo(ctx, lastID); err != nil {
				return err
			}
			time.Sleep(*sleepTime)
		}
	}
	return nil
}

func (agent Agent) checkRepo(ctx context.Context, lastID int64) (int64, error) {
	newestDeployment, err := agent.deployments.findNewestDeployment(ctx, agent.env)
	if err != nil {
		return -1, err
	}

	if newestDeployment == nil || *newestDeployment.ID == lastID {
		log.Printf("No new deployments found.\n")
		return lastID, nil
	}

	log.Printf("Deploying %d\n", *newestDeployment.ID)
	if err := agent.deploy(ctx, newestDeployment); err != nil {
		return -1, err
	}
	return *newestDeployment.ID, nil
}

func (agent Agent) hookContextForDeployment(depl *github.Deployment) hookCtx {
	hookEnv := []string{
		fmt.Sprintf("GITHUB_ENV=%s", agent.env),
		fmt.Sprintf("GITHUB_TASK=%s", *depl.Task),
		fmt.Sprintf("GITHUB_DEPLOYMENT_ID=%d", *depl.ID),
		fmt.Sprintf("GITHUB_DEPLOYMENT_URL=%s", *depl.URL),
		fmt.Sprintf("GITHUB_OWNER=%s", agent.deployments.owner),
		fmt.Sprintf("GITHUB_REPO=%s", agent.deployments.repo),
		fmt.Sprintf("GITHUB_REF=%s", *depl.Ref),
		fmt.Sprintf("GITHUB_SHA=%s", *depl.SHA),
		fmt.Sprintf("GITHUB_CREATOR=%s", *depl.Creator.Login),
		fmt.Sprintf("GITHUB_CREATOR_AVATAR=%s", *depl.Creator.AvatarURL),
	}
	hooks := hookCtx{env: hookEnv}
	return hooks
}

func (agent Agent) deploy(ctx context.Context, depl *github.Deployment) error {
	log.Printf("Starting deployment=%d...\n", *depl.ID)
	if err := agent.deployments.createDeploymentStatus(ctx, depl, "pending", "Firing hook"); err != nil {
		return nil
	}
	hooks := agent.hookContextForDeployment(depl)

	if _, err := hooks.firePreTask(); err != nil {
		log.Printf("pre_task failed: %v\n", err)
	}

	if found, err := hooks.fire(*depl.Task); err != nil {
		if _, err := hooks.firePostFailure(); err != nil {
			log.Printf("post_failure failed: %v\n", err)
		}

		if err := agent.deployments.createDeploymentStatus(ctx, depl, "failure", "Hook failed"); err != nil {
			return nil
		}

	} else if !found {
		log.Printf("No hook '%s' found.\n", *depl.Task)
		if _, err := hooks.firePostFailure(); err != nil {
			log.Printf("post_failure failed: %v\n", err)
		}

		if err := agent.deployments.createDeploymentStatus(ctx, depl, "error", "Unknown hook: "+*depl.Task); err != nil {
			return nil
		}
	} else {
		if _, err := hooks.firePostSuccess(); err != nil {
			log.Printf("post_success failed: %v\n", err)
		}

		if err := agent.deployments.createDeploymentStatus(ctx, depl, "success", "Finished"); err != nil {
			return nil
		}
	}
	return nil
}
