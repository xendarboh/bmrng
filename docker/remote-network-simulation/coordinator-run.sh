#!/bin/bash

set -ex

# create hosts file, generate config from it, launch coordinator using it

args='--numservers 3 --numgroups 3 --numusers 10 --groupsize 3 --numlayers 10'
hostsfile='/src/go/trellis/cmd/experiments/ip.list'

echo -e "server-0\nserver-1\nserver-2" > ${hostsfile}
cd /src/go/0kn/cmd/xtrellis && xtrellis coordinator config ${args} --hostsfile ${hostsfile}
cd /src/go/trellis/cmd/coordinator && coordinator ${args} --runtype 2
