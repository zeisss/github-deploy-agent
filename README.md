# github-deploy-agent

[![wercker status](https://app.wercker.com/status/47137a6deb02cc9038120d2b8e57771b/m "wercker status")](https://app.wercker.com/project/bykey/47137a6deb02cc9038120d2b8e57771b)

The github-deploy-agent is a simple go agent that periodically polls the github api and checks for new deployments.
If found, it executes a hook to deploy.

## Workflow

In my case I use [wercker](https://wercker.com) to create a deployment object in the repository.

The agent periodically polls this repository and looks for the latest deployment.
If a newer than previously deployed exists, it starts deploying it.

 * It creates a deployment status `pending`.
 * A `pre_task` is fired, if available.
 * If tries executing a file called `./hooks/$task`, where `$task` is the string
   from the deployment object. This allows for different actions to be taken.

   The hook gets passed a number of environment variables, including `GITHUB_REF` and `GITHUB_SHA`
   which define which branch/commit of the repository should be commited.

   This hook, which can be anything (bash scripts, binaries etc.) can now perform the actual deploy.

 * The agent creates a `success` status if the exit code is `0`, `failure` otherwise.
 * A `post_success` hook is fired on success with the same environment variables,
   a `post_failure` otherwise. Errors here are ignored.

   This can be used for slack notifications etc.
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

1. On startup, the agent always redeploys the latest deployment found in the repository. It is up
   to the hook to ignore redeploys of the latest version, if wanted.

2. If multiple deployments are created for the same environment while the agent sleeps,
   only the latest is acted upon. If the deployments differ e.g. in the tasks to execute,
   this may lead to unexpected / confusing behavior.

3. Although the agent is single threaded (and thus only applies one deploy at a time), it
   does not help you with downtime-free rollouts etc.

## Future Improvements
* Github Deployment API supports a payload. This is currently not made available to the hook.
* Expose the output of the hooks, e.g. by uploading them to Gist.
