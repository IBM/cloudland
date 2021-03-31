#!/bin/bash

cd `dirname $0`
source ../cloudrc

[ $# -lt 4 ] && echo "$0 <router> <gateway> <mac> <vni> [hard | soft]" && exit -1

router=$1
[ "${router/router-/}" = "$router" ] && router=router-$1
addr=$2
mac=$3
vni=$4
mode=$5

./set_gateway.sh $router $addr $mac $vni $mode
./set_route.sh $router $vni
