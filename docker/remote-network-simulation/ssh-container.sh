#!/bin/bash

source .env

ip=$(docker inspect -f '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' "${1:-coordinator}")
ssh \
  -o StrictHostKeyChecking=no \
  -o UserKnownHostsFile=/dev/null \
  -i volumes/ssh/lkey \
  ${SSHUSER}@${ip}
