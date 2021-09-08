package agent

import (
	"testing"

	"github.com/google/go-github/github"
	"github.com/stretchr/testify/assert"
)

func TestHasState(t *testing.T) {
	var (
		stateInit    = "in_progress"
		stateSuccess = "success"
	)
	statuses := []*github.DeploymentStatus{
		{State: &stateInit},
		{State: &stateSuccess},
	}

	assert.Equal(t, 1, FindState(statuses, DeploymentStateSuccess))
	assert.Equal(t, -1, FindState(statuses, DeploymentStateFailure))
	assert.Equal(t, 1, FindState(statuses, DeploymentStateFailure, DeploymentStateSuccess))
}
