package main

import (
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

func (api deploymentAPI) getDeployments(env string) (deployments []github.Deployment, err error) {
	retry.Do(func() error {
		deployments, _, err = api.client.Repositories.ListDeployments(api.owner, api.repo, &github.DeploymentsListOptions{
			Environment: env,
		})
		return err
	})
	return deployments, err
}

func (api deploymentAPI) findNewestDeployment(env string) (*github.Deployment, error) {
	deployments, err := api.getDeployments(env)
	if err != nil {
		return nil, err
	}

	var newestDeployment github.Deployment
	for _, deployment := range deployments {
		if newestDeployment.CreatedAt == nil || deployment.CreatedAt.Time.After(newestDeployment.CreatedAt.Time) {
			newestDeployment = deployment
		}
	}
	if newestDeployment.ID == nil {
		return nil, nil
	}
	return &newestDeployment, nil
}

// createDeploymentStatus publishes a new status message for the given deployment object.
//
// see https://developer.github.com/v3/repos/deployments/#create-a-deployment-status
// state = pending | success | error | failure
// description = string(140)
func (api deploymentAPI) createDeploymentStatus(depl *github.Deployment, state, desc string) error {
	log.Printf("Setting state=%s descr=%s\n", state, desc)
	return retry.Do(func() error {
		_, _, err := api.client.Repositories.CreateDeploymentStatus(api.owner, api.repo, *depl.ID, &github.DeploymentStatusRequest{
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

func (agent Agent) run() error {
	deployment, err := agent.deployments.findNewestDeployment(agent.env)
	if err != nil {
		return err
	}

	var lastID int
	if deployment != nil {
		log.Printf("Latest deployment is %d. Using as baseline.", *deployment.ID)
		lastID = *deployment.ID
	}
	for {
		if lastID, err = agent.checkRepo(lastID); err != nil {
			return err
		}
		time.Sleep(*sleepTime)
	}
}

func (agent Agent) checkRepo(lastID int) (int, error) {
	newestDeployment, err := agent.deployments.findNewestDeployment(agent.env)
	if err != nil {
		return -1, err
	}

	if newestDeployment == nil || *newestDeployment.ID == lastID {
		log.Printf("No new deployments found.\n")
		return lastID, nil
	}

	log.Printf("Deploying %d\n", *newestDeployment.ID)
	if err := agent.deploy(newestDeployment); err != nil {
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

func (agent Agent) deploy(depl *github.Deployment) error {
	if err := agent.deployments.createDeploymentStatus(depl, "pending", "Firing hook"); err != nil {
		return nil
	}
	hooks := agent.hookContextForDeployment(depl)

	if _, err := hooks.firePreTask(); err != nil {
		log.Printf("pre_task failed: %v\n", err)
	}

	if found, err := hooks.fire(*depl.Task); err != nil {
		if err := agent.deployments.createDeploymentStatus(depl, "error", "Hook failed"); err != nil {
			return nil
		}

		if _, err := hooks.firePostFailure(); err != nil {
			log.Printf("post_failure failed: %v\n", err)
		}
	} else if !found {
		log.Printf("No hook '%s' found.\n", *depl.Task)
		if err := agent.deployments.createDeploymentStatus(depl, "failure", "Unknown hook: "+*depl.Task); err != nil {
			return nil
		}

		if _, err := hooks.firePostFailure(); err != nil {
			log.Printf("post_failure failed: %v\n", err)
		}
	} else {
		if err := agent.deployments.createDeploymentStatus(depl, "success", "Finished"); err != nil {
			return nil
		}

		if _, err := hooks.firePostSuccess(); err != nil {
			log.Printf("post_success failed: %v\n", err)
		}
	}
	return nil
}
