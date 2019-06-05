#!/bin/bash

cd `dirname $0`
source ../cloudrc

[ $# -lt 3 ] && echo "$0 <router> <gateway> <vni> [hard | soft]" && exit -1

router=$1
addr=$2
vni=$3
mode=$4
[ -z "$mode" ] && mode='soft'

bcast=$(ipcalc -b $addr | cut -d= -f2)
./create_link.sh $vni
cat /proc/net/dev | grep -q "^\<ln-$vni\>"
if [ $? -ne 0 ]; then
    ./create_veth.sh $router ln-$vni ns-$vni
fi

if [ "$mode" = "hard" ]; then
    ip netns exec $router ip addr add $addr brd $bcast dev ns-$vni
else
    router_dir=/opt/cloudland/cache/router/$router
    vrrp_conf=$router_dir/keepalived.conf
    pid_file=$router_dir/keepalived.pid
    sed -i "\#$addr dev ns-$vni#d" $vrrp_conf
    sed -i "/virtual_ipaddress {/a $addr dev ns-$vni" $vrrp_conf
    [ -f "$pid_file" ] && ip netns exec $router kill -HUP $(cat $pid_file)
fi
