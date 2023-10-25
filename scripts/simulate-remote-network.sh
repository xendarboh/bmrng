#!/bin/bash

set -ex

make docker-images

# (re)create volume shared by containers
dir=docker/remote-network-simulation
rm -rf ${dir}/volumes
mkdir -m 700 -p ${dir}/volumes/ssh

docker compose \
  --project-directory ${dir} \
  --profile \
  run up
