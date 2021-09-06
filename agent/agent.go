package agent

import (
	"context"
	"time"
)

// Agent applies deployments
type Agent struct {
	Log         Logger
	Deployments *DeploymentOptions
	Deployer    *Deployer
}

func (agent Agent) Run(ctx context.Context, loop bool, sleepTime time.Duration) error {
	deployment, err := agent.Deployments.FindNewestDeployment(ctx)
	if err != nil {
		return err
	}

	var lastID int64
	if deployment != nil {
		// If the latest deployment has no success message, deploy it immediately
		if success, err := agent.Deployments.HasSuccessStatus(ctx, deployment); err != nil {
			return err
		} else if !success {
			agent.Log.Printf("Found new deployment=%d\n", *deployment.ID)
			if err := agent.Deployer.Deploy(ctx, deployment); err != nil {
				return err
			}
		} else {
			agent.Log.Printf("Latest deployment %d (from %s) has 'success' message. Using as baseline.", *deployment.ID, deployment.CreatedAt)
		}
		lastID = *deployment.ID
	} else {
		agent.Log.Println("No deployment in repository found.")
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
	newestDeployment, err := agent.Deployments.FindNewestDeployment(ctx)
	if err != nil {
		return -1, err
	}

	if newestDeployment == nil || *newestDeployment.ID == lastID {
		agent.Log.Printf("No new deployments found.\n")
		return lastID, nil
	}

	agent.Log.Printf("Deploying %d\n", *newestDeployment.ID)
	if err := agent.Deployer.Deploy(ctx, newestDeployment); err != nil {
		return -1, err
	}
	return *newestDeployment.ID, nil
}
