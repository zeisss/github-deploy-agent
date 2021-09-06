package agent

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHooks(t *testing.T) {
	hooks := givenTestdataHooks()
	hooks.env = []string{
		"HELLO=WORLD",
	}

	err := hooks.fireCustom("helloworld.bash")
	assert.NoError(t, err)

	err = hooks.fireCustom("print_env.bash")
	assert.NoError(t, err)
}

func TestHooks__NonZeroExitCodeReturnsError(t *testing.T) {
	h := givenTestdataHooks()

	err := h.fireCustom("exit_error.bash")
	assert.Error(t, err)
}

func TestHooks__NonExistingHookReturnsErrHookNotFound(t *testing.T) {
	h := givenTestdataHooks()

	err := h.fireCustom("some_hook_filename_that_does_not_exist")
	assert.ErrorIs(t, err, ErrHookNotFound)
}

func givenTestdataHooks() *Hooks {
	return &Hooks{
		Path: "../testdata",
	}
}
