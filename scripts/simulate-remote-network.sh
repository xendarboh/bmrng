#!/bin/bash

set -ex

make docker-images

# (re)create volume shared by containers
dir_volumes=docker/remote-network-simulation/volumes
rm -rf ${dir_volumes}
mkdir -m 700 -p ${dir_volumes}/ssh

cd docker/remote-network-simulation && docker compose --profile run up
