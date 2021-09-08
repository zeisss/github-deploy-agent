package agent

import (
	"context"

	"github.com/giantswarm/retry-go"
	"github.com/google/go-github/github"
)

type deploymentState string

const (
	DeploymentStateError   deploymentState = "error"
	DeploymentStateFailure deploymentState = "failure"
	DeploymentStateSuccess deploymentState = "success"
)

type DeploymentOptions struct {
	Owner, Repo string
	Env         string
	Client      *github.Client
}

func (api *DeploymentOptions) ListDeployments(ctx context.Context) (deployments []*github.Deployment, err error) {
	retry.Do(func() error {
		deployments, _, err = api.Client.Repositories.ListDeployments(ctx, api.Owner, api.Repo, &github.DeploymentsListOptions{
			Environment: api.Env,
		})
		return err
	})
	return deployments, err
}

func (api *DeploymentOptions) FindNewestDeployment(ctx context.Context) (*github.Deployment, error) {
	deployments, err := api.ListDeployments(ctx)
	if err != nil {
		return nil, err
	}

	var newestDeployment *github.Deployment
	for _, deployment := range deployments {
		if newestDeployment == nil || deployment.CreatedAt.Time.After(newestDeployment.CreatedAt.Time) {
			newestDeployment = deployment
		}
	}
	if newestDeployment == nil {
		return nil, nil
	}
	return newestDeployment, nil
}

func (api *DeploymentOptions) HasSuccessStatus(ctx context.Context, depl *github.Deployment) (bool, error) {
	statuses, _, err := api.Client.Repositories.ListDeploymentStatuses(ctx, api.Owner, api.Repo, *depl.ID, &github.ListOptions{})
	if err != nil {
		return false, err
	}
	return FindState(statuses, DeploymentStateSuccess) >= 0, nil
}

func FindState(statuses []*github.DeploymentStatus, needleStates ...deploymentState) int {
	for index, status := range statuses {
		for _, needle := range needleStates {
			if status.State != nil && *status.State == string(needle) {
				return index
			}
		}
	}
	return -1
}

// CreateDeploymentStatus publishes a new status message for the given deployment object.
//
// see https://developer.github.com/v3/repos/deployments/#create-a-deployment-status
// state = pending | success | error | failure
// description = string(140)
func (api *DeploymentOptions) CreateDeploymentStatus(ctx context.Context, depl *github.Deployment, state deploymentState, desc string) error {
	s := string(state)
	return retry.Do(func() error {
		_, _, err := api.Client.Repositories.CreateDeploymentStatus(ctx, api.Owner, api.Repo, *depl.ID, &github.DeploymentStatusRequest{
			State:       &s,
			Description: &desc,
		})
		return err
	})
}
