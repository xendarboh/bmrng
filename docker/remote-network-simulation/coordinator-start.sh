#!/bin/bash

set -ex

# generate an ssh key for coordinator host to ssh to simulated remote servers

dir="/home/${SSHUSER}/.ssh"

# NOTE 20231020: ~/.ssh/lkey hardcoded in trellis
test ! -f ${dir}/lkey \
  && ssh-keygen \
    -t ed25519 \
    -a 100 \
    -f ${dir}/lkey \
    -N ""

cat ${dir}/lkey.pub > ${dir}/authorized_keys

chown -R ${SSHUSER}:${SSHUSER} ${dir}

# wait for servers to start, then run a coordinated mix-net experiment
sleep 3s && su -c '/src/docker/remote-network-simulation/coordinator-run.sh' ${SSHUSER} &

# start ssh daemon, useful for local dev
/usr/sbin/sshd -D
