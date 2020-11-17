#!/bin/bash

cd $(dirname $0)
source ../cloudrc

[ $# -lt 3 ] && echo "$0 <router> <veth_name> <peer_name>" && exit -1
router=$1
device=$2
peerdev=$3

ip link add $device type veth peer name $peerdev
ip link set $device up
ip link set $peerdev netns $router
ip netns exec $router ip link set $peerdev mtu 1450 up
prefix=${device%%-*}
if [ "$prefix" == "ext" ]; then
    ./create_link.sh $external_vlan
    bridge=br$external_vlan
elif [ "$prefix" == "int" ]; then
    ./create_link.sh $internal_vlan
    bridge=br$internal_vlan
elif [ "$prefix" == "ln" ]; then
    vni=${device##*-}
    bridge=br$vni
fi
ip link set dev $device master $bridge
