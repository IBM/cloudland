#!/bin/bash

cd $(dirname $0)
source ../cloudrc

[ $# -lt 1 ] && echo "$0 <router>" && exit -1

ID=$1
router=router-$ID

[ -z "$router" ] && exit 1

ip netns exec $router ip link show | grep ns-
[ $? -eq 0 ] && echo "Active subnet gateways exist" && exit 0
ip netns exec $router ip link set lo down

suffix=$1
ip link del ext-$suffix
ip link del int-$suffix
ip netns del $router