#!/bin/bash

set -ex

xtrellis \
  coordinator \
  mixnet \
  --gatewayenable \
  --debug \
  ${@} \
  &

xtrellis_pid=$!

sleep 10s

./scripts/test-gateway-io.sh 102400

# kill spawned mix-net servers
pkill -P ${xtrellis_pid}

# kill the coordinator (and don't care if kill fails)
kill ${xtrellis_pid} || exit 0
