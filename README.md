# github-deploy-agent

The github-deploy-agent is a simple go agent that periodically polls the github api and checks for new deployments.
If found, it executes a hook to deploy.

## Workflow

In my case I use [wercker](https://wercker.com) to create a deployment object in the repository.

On startup the agent looks for the latest deployment object, and deploys it, if no
success message has been created yet.

The agent then periodically polls the repository and grabs the latest deployment.
If it is newer than the previously deployed one, it starts deploying the new one.

* It creates a deployment status `pending`.
* A `pre_task` is fired, if available.
* If tries executing a file called `./hooks/$task`, where `$task` is the string
  from the deployment object. This allows for different actions to be taken.

  The hook gets passed a number of environment variables (see below)
  which define which branch/commit of the repository should be committed.

   This hook, which can be anything (bash scripts, binaries etc.) can now perform the actual deploy.
* A `post_success` hook is fired on success with the same environment variables,
  a `post_failure` otherwise. Errors here are ignored.

   This can be used for slack notifications etc.
* The agent creates a `success` status if the exit code is `0`, `failure` otherwise.
  If the task didn't exist as a hook, an `error` status is raised.

* The agent sleeps for `--sleep` before checking for the next deployment again.

## Environment Variables

All hooks are passed the same environment variables.

Name                   | Desc
-----------------------|--------------------
GITHUB_TASK            | The task name
GITHUB_DEPLOYMENT_ID   | ID of the deployment
GITHUB_DEPLOYMENT_URL  | URL of the deployment object in the Github API
GITHUB_REF             | Branch name to be deployed
GITHUB_SHA             | SHA of the commit to be deployed
GITHUB_OWNER           | Owner of the repository
GITHUB_REPO            | Repository name
GITHUB_CREATOR         | Name of the user that created the deployment
GITHUB_CREATOR_AVATAR  | Avatar URL of the creator
GITHUB_ENV             | The env the agent is listening on

## Caveats

1. If multiple deployments are created for the same environment while the agent sleeps,
   only the latest is acted upon. If the deployments differ e.g. in the tasks to execute,
   this may lead to unexpected / confusing behavior.

2. Although the agent is single threaded (and thus only applies one deploy at a time), it
   does not help you with downtime-free rollouts etc.

## Future Improvements

* Github Deployment API supports a payload. This is currently not made available to the hook.
* Expose the output of the hooks, e.g. by uploading them to Gist.
