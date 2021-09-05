package agent

import (
	"testing"

	"github.com/google/go-github/github"
	"github.com/stretchr/testify/assert"
)

func TestHasState(t *testing.T) {
	var (
		stateInit    = "init"
		stateSuccess = "success"
	)
	statuses := []*github.DeploymentStatus{
		{State: &stateInit},
		{State: &stateSuccess},
	}

	assert.True(t, hasState(statuses, "success"))
	assert.False(t, hasState(statuses, "errror"))
}
