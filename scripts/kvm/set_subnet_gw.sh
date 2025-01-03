#!/bin/bash

cd `dirname $0`
source ../cloudrc

[ $# -lt 3 ] && echo "$0 <router> <vlan> <gateway>" && exit -1

router=$1
[ "${router/router-/}" = "$router" ] && router=router-$1
vlan=$2
gateway=$3

[ "$router" = "router-0" ] && exit 0

./create_local_router.sh $router
cat /proc/net/dev | grep -q "^\<ln-$vlan\>"
if [ $? -ne 0 ]; then
    ./create_veth.sh $router ln-$vlan ns-$vlan
    apply_vnic -I ln-$vlan
fi
brctl addif br$vlan ln-$vlan
ip_net=$(ipcalc -b $gateway | grep Network | awk '{print $2}')
ip netns exec $router ipset add nonat $ip_net
bcast=$(ipcalc -b $gateway | grep Broadcast | awk '{print $2}')
ip netns exec $router ip addr add $gateway brd $bcast dev ns-$vlan
mac_map=$(printf "%06x" $vlan)
hw_addr=52:$(echo $mac_map | cut -c 1-2):$(echo $mac_map | cut -c 3-4):$(echo $mac_map | cut -c 5-6)
hyper_map=$(printf "%04x" $(($SCI_CLIENT_ID & 0xffff)))
hw_addr=$hw_addr:$(echo $hyper_map | cut -c 1-2):$(echo $hyper_map | cut -c 3-4)
if [ $? -eq 0 ]; then
    ip netns exec $router ip link set ns-$vlan address $hw_addr
fi
link_mtu=$(ip -o link show $vxlan_interface | sed "s/.* mtu \(.*\) qdisc.*/\1/")
if [ $link_mtu -gt 1400 ]; then
    link_mtu=$(( $link_mtu - 50 ))
    ip netns exec $router ip link set ns-$vlan mtu $link_mtu
fi
