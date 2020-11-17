#!/bin/bash

cd `dirname $0`
source ../cloudrc

[ $# -lt 3 ] && echo "$0 <router> <gateway> <vni> [hard | soft]" && exit -1

router=$1
[ "${router/router-/}" = "$router" ] && router=router-$1
addr=$2
vni=$3
mode=$4

./set_gateway.sh $router $addr $vni $mode
./set_route.sh $router $vni
