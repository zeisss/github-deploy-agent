package agent

import (
	"fmt"
	"log"
	"os"
	"os/exec"
)

// hooksCtx is used to define how to execute the hook scripts.
// When creating, the environment variables are provided.
type Hooks struct {
	env []string
}

func (h Hooks) _fire(name string) (bool, error) {
	_, err := os.Stat("./hooks/" + name)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}

	log.Printf("Firing hook %s\n", name)
	cmd := exec.Command("./hooks/" + name)
	cmd.Env = h.env
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return true, cmd.Run()
}

func (h Hooks) fire(name string) (bool, error) {
	if name == "post_success" || name == "post_failure" || name == "pre_task" {
		return false, fmt.Errorf("reserved hook name: %s", name)
	}
	return h._fire(name)
}

func (h Hooks) firePostSuccess() (bool, error) {
	return h._fire("post_success")
}

func (h Hooks) firePostFailure() (bool, error) {
	return h._fire("post_failure")
}

func (h Hooks) firePreTask() (bool, error) {
	return h._fire("pre_task")
}
