package agent

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/go-github/github"
)

// Agent applies deployments
type Agent struct {
	Env         string
	Deployments *DeploymentAPI
}

func (agent Agent) Run(ctx context.Context, loop bool, sleepTime time.Duration) error {
	deployment, err := agent.Deployments.findNewestDeployment(ctx, agent.Env)
	if err != nil {
		return err
	}

	var lastID int64
	if deployment != nil {
		// If the latest deployment has no success message, deploy it immediately
		if success, err := agent.Deployments.hasSuccessStatus(ctx, deployment); err != nil {
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
			time.Sleep(sleepTime)
		}
	}
	return nil
}

func (agent Agent) checkRepo(ctx context.Context, lastID int64) (int64, error) {
	newestDeployment, err := agent.Deployments.findNewestDeployment(ctx, agent.Env)
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

func (agent Agent) hookContextForDeployment(depl *github.Deployment) Hooks {
	hookEnv := []string{
		fmt.Sprintf("GITHUB_ENV=%s", agent.Env),
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

func (agent Agent) deploy(ctx context.Context, depl *github.Deployment) error {
	log.Printf("Starting deployment=%d...\n", *depl.ID)
	if err := agent.Deployments.createDeploymentStatus(ctx, depl, "pending", "Firing hook"); err != nil {
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

		if err := agent.Deployments.createDeploymentStatus(ctx, depl, "failure", "Hook failed"); err != nil {
			return nil
		}

	} else if !found {
		log.Printf("No hook '%s' found.\n", *depl.Task)
		if _, err := hooks.firePostFailure(); err != nil {
			log.Printf("post_failure failed: %v\n", err)
		}

		if err := agent.Deployments.createDeploymentStatus(ctx, depl, "error", "Unknown hook: "+*depl.Task); err != nil {
			return nil
		}
	} else {
		if _, err := hooks.firePostSuccess(); err != nil {
			log.Printf("post_success failed: %v\n", err)
		}

		if err := agent.Deployments.createDeploymentStatus(ctx, depl, "success", "Finished"); err != nil {
			return nil
		}
	}
	return nil
}
