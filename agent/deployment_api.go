package agent

import (
	"context"
	"strings"

	"github.com/giantswarm/retry-go"
	"github.com/google/go-github/github"
)

type deploymentState string

const (
	deploymentStateError      deploymentState = "error"
	deploymentStateFailure    deploymentState = "failure"
	deploymentStatePending    deploymentState = "pending"
	deploymentStateInProgress deploymentState = "in_progress"
	deploymentStateQueued     deploymentState = "queued"
	deploymentStateSuccess    deploymentState = "success"
)

func NewDeploymentAPI(repository, env string, client *github.Client) *DeploymentOptions {
	s := strings.Split(repository, "/")
	owner := s[0]
	repo := s[1]

	return &DeploymentOptions{
		owner:  owner,
		repo:   repo,
		env:    env,
		client: client,
	}
}

type DeploymentOptions struct {
	owner, repo string
	env         string
	client      *github.Client
}

func (api *DeploymentOptions) ListDeployments(ctx context.Context) (deployments []*github.Deployment, err error) {
	retry.Do(func() error {
		deployments, _, err = api.client.Repositories.ListDeployments(ctx, api.owner, api.repo, &github.DeploymentsListOptions{
			Environment: api.env,
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
	statuses, _, err := api.client.Repositories.ListDeploymentStatuses(ctx, api.owner, api.repo, *depl.ID, &github.ListOptions{})
	if err != nil {
		return false, err
	}
	return hasState(statuses, "success"), nil
}

func hasState(statuses []*github.DeploymentStatus, needleState deploymentState) bool {
	for _, status := range statuses {
		if status.State != nil && *status.State == string(needleState) {
			return true
		}
	}
	return false
}

// CreateDeploymentStatus publishes a new status message for the given deployment object.
//
// see https://developer.github.com/v3/repos/deployments/#create-a-deployment-status
// state = pending | success | error | failure
// description = string(140)
func (api *DeploymentOptions) CreateDeploymentStatus(ctx context.Context, depl *github.Deployment, state deploymentState, desc string) error {
	s := string(state)
	return retry.Do(func() error {
		_, _, err := api.client.Repositories.CreateDeploymentStatus(ctx, api.owner, api.repo, *depl.ID, &github.DeploymentStatusRequest{
			State:       &s,
			Description: &desc,
		})
		return err
	})
}
