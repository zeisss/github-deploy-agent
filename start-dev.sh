#!/bin/bash

set -eu
test ! -z "${GITHUB_TOKEN}"

## Create a tmp folder to run this in
rm -rf tmp
mkdir tmp
cd tmp

## Build some noisy hooks
mkdir hooks
printf '#!/bin/sh\necho pre_task hook\nenv' > hooks/pre_task
printf '#!/bin/sh\necho post_success hook\nenv' > hooks/post_success
printf '#!/bin/sh\necho post_failure hook\nenv' > hooks/post_failure
printf '#!/bin/sh\necho deploy hook\nenv' > hooks/deploy
chmod +x hooks/*

## Execute the agent
../github-deploy-agent server \
  --env="testing" \
  --repository="zeisss/github-deploy-agent" \
  --token="${GITHUB_TOKEN}" \
  --sleep="30s"

## To create deployments
# curl -XPOST https://api.github.com/repos/zeisss/github-deploy-agent/deployments \
#   -H"Authorization: bearer ${GITHUB_TOKEN}" \
#   -d '{"environment":"testing", "ref":"master", "required_contexts": [], "task": "deploy"}'
