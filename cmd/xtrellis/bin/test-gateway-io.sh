#!/bin/bash

# Test transmission of data streams through mix-net by means of gateway proxy i/o

# Example coordinated local mix-net invocation before running this:
# ./xtrellis coordinator --messagesize 120 --numlayers 5 --numusers 3 --roundinterval 1 --gatewayenable --debug

set -e

# bytes of data to send through the mix-net (default 10K)
DATA_SIZE=${1:-10240}

cd /tmp
rm -f data.{in,out}

# random -> file
cat /dev/urandom | head -c ${DATA_SIZE} > data.in
# dd if=/dev/urandom of=data.in bs=1024 count=10 &>/dev/null

# file -> gateway -> [mix-net]
cat data.in | nc -q 1 localhost 9000

# [mix-net] -> gateway -> file
wget -O data.out -q http://localhost:9900 || echo "not found"

# compare data input to output
if diff -q data.in data.out &>/dev/null; then
  echo "Success!"
else
  echo "Fail!"
fi
