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
read -d'\n' -r network bcast hostmin hostmax < <(ipcalc -nb $gateway | awk '/Network/ {print $2} /Broadcast/ {print $2} /HostMin/ {print $2} /HostMax/ {print $2}')
ip netns exec $router ipset add nonat $network
ip netns exec $router ip addr add $gateway brd $bcast dev ns-$vlan
mac_map=$(printf "%06x" $vlan)
hw_addr=52:$(echo $mac_map | cut -c 1-2):$(echo $mac_map | cut -c 3-4):$(echo $mac_map | cut -c 5-6)
hyper_map=$(printf "%04x" $(($SCI_CLIENT_ID & 0xffff)))
hw_addr=$hw_addr:$(echo $hyper_map | cut -c 1-2):$(echo $hyper_map | cut -c 3-4)
if [ $? -eq 0 ]; then
    ip netns exec $router ip link set ns-$vlan address $hw_addr
fi
./set_subnet_dhcp.sh "$router" "$vlan" "$gateway" "$network" "$hostmin" "$hostmax"
