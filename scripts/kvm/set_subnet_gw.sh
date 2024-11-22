#!/bin/bash

cd `dirname $0`
source ../cloudrc

[ $# -lt 4 ] && echo "$0 <router> <vlan> <gateway> <ext_vlan>" && exit -1

router=$1
[ "${router/router-/}" = "$router" ] && router=router-$1
vlan=$2
gateway=$3
ext_vlan=$4

./create_local_router.sh $router $ext_vlan
cat /proc/net/dev | grep -q "^\<ln-$vni\>"
if [ $? -ne 0 ]; then
    ./create_veth.sh $router ln-$vlan ns-$vlan
    apply_vnic -I ln-$vni
fi
brctl addif br$vlan ln-$vlan
ip_net=$(ipcalc -b $gateway | grep Network | awk '{print $2}')
ip netns exec $router ipset add nonat $ip_net
bcast=$(ipcalc -b $gateway | grep Broadcast | awk '{print $2}')
ip netns exec $router ip addr add $gateway brd $bcast dev ns-$vlan
ip netns exec $router ip route add $ip_net dev ns-$vlan table fip
mac_map=$(printf "%06x" $vlan)
hw_addr=52:$(echo $mac_map | cut -c 1-2):$(echo $mac_map | cut -c 3-4):$(echo $mac_map | cut -c 5-6)
hyper_map=$(printf "%04x" $(($SCI_CLIENT_ID & 0xffff)))
hw_addr=$hw_addr:$(echo $hyper_map | cut -c 1-2):$(echo $hyper_map | cut -c 3-4)
if [ $? -eq 0 ]; then
    ip netns exec $router ip link set ns-$vlan address $hw_addr
fi
