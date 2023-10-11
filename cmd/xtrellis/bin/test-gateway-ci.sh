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

kill -s SIGTERM ${xtrellis_pid}

exit 0
