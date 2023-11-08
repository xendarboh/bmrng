#!/bin/bash

set -ex

export _0KN_WORKDIR=~/.0KN
mkdir -p ${_0KN_WORKDIR}

# create hosts file, generate config from it, launch coordinator using it

args='--numservers 3 --numgroups 3 --numusers 10 --groupsize 3 --numlayers 10'
hostsfile="${_0KN_WORKDIR}/ip.list"

echo -e "server-0\nserver-1\nserver-2" > ${hostsfile}
xtrellis coordinator config ${args} --hostsfile ${hostsfile}
xtrellis coordinator experiment ${args} --networktype 2
