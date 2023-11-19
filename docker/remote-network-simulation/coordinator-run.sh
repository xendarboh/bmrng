#!/bin/bash

set -ex

export _0KN_WORKDIR=~/.0KN
mkdir -p ${_0KN_WORKDIR}

# minimal mix-net args
args='--numservers 3 --numgroups 3 --numusers 10 --groupsize 3 --numlayers 10'

# create hosts file for minimal mix-net
hostsfile="${_0KN_WORKDIR}/hosts.list"
echo -e "server-0:8000\nserver-1:8000\nserver-2:8000" > ${hostsfile}

# generate mix-net round config from hosts file, share with each host
xtrellis coordinator config ${args} --hostsfile ${hostsfile}

# launch coordinated experiment using remote hosts
xtrellis coordinator experiment ${args} --networktype 2

# test gateway using remote hosts
cd /src && ./scripts/test-gateway-ci.sh ${args} --networktype 2
