#!/bin/bash

cd `dirname $0`
source ../cloudrc

[ $# -lt 1 ] && die "$0 <lxname>"

lxname=$1
bridge=$(lxc config device get $lxname eth0 parent )
[ -n "$bridge" ] && vni=${bridge#br}
lxc delete $lxname --force
./clear_link.sh $vni
sidecar span log $span "Callback: `basename $0` '$ID'"
echo "|:-COMMAND-:| `basename $0` '$ID'"
