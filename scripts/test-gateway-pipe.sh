#!/bin/bash

# Test transmission of data streams through mix-net by means of gateway proxy i/o
# Example: cat in.png | ./test-gateway-pipe.sh > out.png

set -e

if [ -p /dev/stdin ]; then
  cat | nc -q 1 localhost 9000
else
  echo "No input given!"
fi

wget -O - -q http://localhost:9900 || echo "Not found"
