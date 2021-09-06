package agent

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

var ErrHookNotFound error = errors.New("unknown hook")

// hooksCtx is used to define how to execute the hook scripts.
// When creating, the environment variables are provided.
type Hooks struct {
	Path string
	env  []string
}

func (h Hooks) _fire(name string) error {
	_, err := os.Stat(filepath.Join(h.Path, name))
	if err != nil {
		if os.IsNotExist(err) {
			return ErrHookNotFound
		}
		return err
	}

	log.Printf("Firing hook %s\n", name)
	cmd := exec.Command(filepath.Join(h.Path, name))
	cmd.Env = h.env
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("hook error: %w", err)
	}
	return nil
}

func (h Hooks) fireCustom(name string) error {
	if name == "post_success" || name == "post_failure" || name == "pre_task" {
		return fmt.Errorf("reserved hook name: %s", name)
	}
	return h._fire(name)
}

func (h Hooks) firePostSuccess() error {
	return h._fire("post_success")
}

func (h Hooks) firePostFailure() error {
	return h._fire("post_failure")
}

func (h Hooks) firePreTask() error {
	return h._fire("pre_task")
}
