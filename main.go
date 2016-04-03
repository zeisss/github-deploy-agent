package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/giantswarm/retry-go"
	"github.com/google/go-github/github"
	"github.com/spf13/pflag"
	"golang.org/x/oauth2"
)

var (
	repository = pflag.String("repository", "", "The repository to deploy for - owner/repo")
	env        = pflag.String("env", "production", "The environment to act on")
	token      = pflag.String("token", "", "API token for github")
	sleepTime  = pflag.Duration("sleep", 60*time.Second, "Sleep time between checks to github api")
)

func main() {
	pflag.Parse()

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: *token},
	)
	tc := oauth2.NewClient(oauth2.NoContext, ts)

	client := github.NewClient(tc)

	s := strings.Split(*repository, "/")
	owner := s[0]
	repo := s[1]

	deployment, err := findNewestDeployment(client, owner, repo)
	if err != nil {
		panic(err.Error())
	}

	var lastID int
	if deployment != nil {
		log.Printf("Latest deployment is %d. Using as baseline.", *deployment.ID)
		lastID = *deployment.ID
	}
	for {
		if lastID, err = checkRepo(client, owner, repo, lastID); err != nil {
			panic(err.Error())
		}
		time.Sleep(*sleepTime)
	}
}

func checkRepo(client *github.Client, owner, repo string, lastID int) (int, error) {
	newestDeployment, err := findNewestDeployment(client, owner, repo)
	if err != nil {
		return -1, err
	}

	if newestDeployment == nil || *newestDeployment.ID == lastID {
		log.Printf("No new deployments found.\n")
		return lastID, nil
	}

	log.Printf("Deploying %d\n", *newestDeployment.ID)
	if err := deploy(client, owner, repo, newestDeployment); err != nil {
		return -1, err
	}
	return *newestDeployment.ID, nil
}

func deploy(client *github.Client, owner, repo string, depl *github.Deployment) error {
	if err := createDeploymentStatus(client, owner, repo, depl, "pending", "Firing deployment hook"); err != nil {
		return nil
	}

	env := []string{
		fmt.Sprintf("GITHUB_ENV=%s", *env),
		fmt.Sprintf("GITHUB_TASK=%s", *depl.Task),
		fmt.Sprintf("GITHUB_DEPLOYMENT_ID=%d", *depl.ID),
		fmt.Sprintf("GITHUB_DEPLOYMENT_URL=%s", *depl.URL),
		fmt.Sprintf("GITHUB_OWNER=%s", owner),
		fmt.Sprintf("GITHUB_REPO=%s", repo),
		fmt.Sprintf("GITHUB_REF=%s", *depl.Ref),
		fmt.Sprintf("GITHUB_SHA=%s", *depl.SHA),
		fmt.Sprintf("GITHUB_CREATOR=%s", *depl.Creator.Login),
		fmt.Sprintf("GITHUB_CREATOR_AVATAR=%s", *depl.Creator.AvatarURL),
	}

	if _, err := fireHook("pre_task", env); err != nil {
		log.Printf("pre_task failed: %v\n", err)
	}

	if found, err := fireHook(*depl.Task, env); err != nil {
		if err := createDeploymentStatus(client, owner, repo, depl, "failure", "Hook failed"); err != nil {
			return nil
		}

		if _, err := fireHook("post_failure", env); err != nil {
			log.Printf("post_failure failed: %v\n", err)
		}
	} else if !found {
		log.Printf("No hook '%s' found.\n", *depl.Task)
		if err := createDeploymentStatus(client, owner, repo, depl, "failure", "Unknown hook: "+*depl.Task); err != nil {
			return nil
		}
	} else {
		if err := createDeploymentStatus(client, owner, repo, depl, "success", "Finished"); err != nil {
			return nil
		}

		if _, err := fireHook("post_success", env); err != nil {
			log.Printf("post_success failed: %v\n", err)
		}
	}
	return nil
}

func createDeploymentStatus(client *github.Client, owner, repo string, depl *github.Deployment, state, desc string) error {
	log.Printf("Setting state=%s descr=%s\n", state, desc)
	return retry.Do(func() error {
		_, _, err := client.Repositories.CreateDeploymentStatus(owner, repo, *depl.ID, &github.DeploymentStatusRequest{
			State:       &state,
			Description: &desc,
		})
		return err
	})
}

func getDeployments(client *github.Client, owner, repo, env string) (deployments []github.Deployment, err error) {
	retry.Do(func() error {
		deployments, _, err = client.Repositories.ListDeployments(owner, repo, &github.DeploymentsListOptions{
			Environment: env,
		})
		return err
	})
	return deployments, err
}

func findNewestDeployment(client *github.Client, owner, repo string) (*github.Deployment, error) {
	deployments, err := getDeployments(client, owner, repo, *env)
	if err != nil {
		return nil, err
	}

	var newestDeployment github.Deployment
	for _, deployment := range deployments {
		if newestDeployment.CreatedAt == nil || deployment.CreatedAt.Time.After(newestDeployment.CreatedAt.Time) {
			newestDeployment = deployment
		}
	}
	if newestDeployment.ID == nil {
		return nil, nil
	}
	return &newestDeployment, nil
}

func fireHook(name string, env []string) (bool, error) {
	_, err := os.Stat("./hooks/" + name)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}

	log.Printf("Firing hook %s\n", name)
	cmd := exec.Command("./hooks/" + name)
	cmd.Env = env
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return true, cmd.Run()
}
