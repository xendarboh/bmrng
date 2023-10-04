#!/bin/bash

# script's directory for invocation regardless of pwd
script_dir=$(readlink -f $(dirname $0))

${script_dir}/../xtrellis \
  coordinator \
  --messagesize 120 \
  --numlayers 5 \
  --numusers 3 \
  --roundinterval 1 \
  --gatewayenable \
  --debug
