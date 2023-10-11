#!/bin/bash

set -ex

./xtrellis \
  coordinator \
  --gatewayenable \
  --debug \
  &

xtrellis_pid=$!

sleep 10s

./bin/test-gateway-io.sh 102400

# kill spawned mix-net servers
pkill -P ${xtrellis_pid}

# kill the coordinator
kill ${xtrellis_pid}

exit 0
