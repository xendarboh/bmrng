#!/bin/bash

# Test transmission of data streams through mix-net by means of gateway proxy i/o

# First start a coordinated local mix-net with gateway enabled.
# Run the script, or set options manually: ./bin/run-coordinator-gateway.sh

set -e

# bytes of data to send through the mix-net (default 10K)
DATA_SIZE=${1:-10240}

GATEWAY_HOST=${2:-localhost}
GATEWAY_PORT_IN=${3:-9000}
GATEWAY_PORT_OUT=${4:-9900}

cd /tmp
rm -f data.{in,out}

# random -> file
cat /dev/urandom | head -c ${DATA_SIZE} > data.in
# dd if=/dev/urandom of=data.in bs=1024 count=10 &>/dev/null

# file -> gateway -> [mix-net]
cat data.in | nc -q 1 ${GATEWAY_HOST} ${GATEWAY_PORT_IN}

# [mix-net] -> gateway -> file
wget -O data.out -q http://${GATEWAY_HOST}:${GATEWAY_PORT_OUT} || echo "not found"

# compare data input to output
if diff -q data.in data.out &>/dev/null; then
  echo "Success!"
  exit 0
else
  echo "Fail!"
  exit 1
fi
