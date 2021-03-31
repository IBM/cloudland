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
[ -z "$mode" ] && mode='soft'

bcast=$(ipcalc -b $addr | cut -d= -f2)
./create_link.sh $vni
cat /proc/net/dev | grep -q "^\<ln-$vni\>"
if [ $? -ne 0 ]; then
    ./create_veth.sh $router ln-$vni ns-$vni
fi
apply_vnic -I ln-$vni

iface=ns-$vni
router_dir=/opt/cloudland/cache/router/$router
vrrp_conf=$router_dir/keepalived.conf
pid_file=$router_dir/keepalived.pid
sed -i "\#.* dev $iface#d" $vrrp_conf
#addrs=$(ip netns exec $router ip addr show $iface | grep 'inet ' | awk '{print $2}')
#for addr in $addrs; do
#    ip netns exec $router ip addr del $addr dev $iface
#done

ip netns exec $router ip link set $iface address $mac
if [ "$mode" = "hard" ]; then
    ip netns exec $router ip addr add $addr brd $bcast dev $iface
else
    sed -i "/virtual_ipaddress {/a $addr dev $iface" $vrrp_conf
    ip netns exec $router ipset add nonat $addr
    [ -f "$pid_file" ] && ip netns exec $router kill -HUP $(cat $pid_file)
fi
