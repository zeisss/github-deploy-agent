#!/bin/bash

set -eu
test ! -z "${GITHUB_PERSONAL_ACCESS_TOKEN}"

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
../github-deploy-agent \
  --repository="zeisss/github-deploy-agent" \
  --token="${GITHUB_PERSONAL_ACCESS_TOKEN}" \
  --sleep="30s"
