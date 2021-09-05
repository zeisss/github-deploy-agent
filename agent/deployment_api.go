package agent

import (
	"context"
	"log"
	"strings"

	"github.com/giantswarm/retry-go"
	"github.com/google/go-github/github"
)

func NewDeploymentAPI(repository string, client *github.Client) *DeploymentAPI {
	s := strings.Split(repository, "/")
	owner := s[0]
	repo := s[1]

	return &DeploymentAPI{
		owner:  owner,
		repo:   repo,
		client: client,
	}
}

type DeploymentAPI struct {
	owner  string
	repo   string
	client *github.Client
}

func (api *DeploymentAPI) getDeployments(ctx context.Context, env string) (deployments []*github.Deployment, err error) {
	retry.Do(func() error {
		deployments, _, err = api.client.Repositories.ListDeployments(ctx, api.owner, api.repo, &github.DeploymentsListOptions{
			Environment: env,
		})
		return err
	})
	return deployments, err
}

func (api *DeploymentAPI) findNewestDeployment(ctx context.Context, env string) (*github.Deployment, error) {
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

func (api *DeploymentAPI) hasSuccessStatus(ctx context.Context, depl *github.Deployment) (bool, error) {
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
func (api *DeploymentAPI) createDeploymentStatus(ctx context.Context, depl *github.Deployment, state, desc string) error {
	log.Printf("Setting state=%s descr=%s\n", state, desc)
	return retry.Do(func() error {
		_, _, err := api.client.Repositories.CreateDeploymentStatus(ctx, api.owner, api.repo, *depl.ID, &github.DeploymentStatusRequest{
			State:       &state,
			Description: &desc,
		})
		return err
	})
}
