#!/bin/bash

cd `dirname $0`
source ../cloudrc

[ $# -lt 3 ] && echo "$0 <router> <gateway> <vni>" && exit -1

router=$1
addr=$2
vni=$3

bcast=$(ipcalc -b $addr | cut -d= -f2)
router_dir=/opt/cloudland/cache/router/$router
vrrp_conf=$router_dir/keepalived.conf
pid_file=$router_dir/keepalived.pid
sed -i "\#$addr dev ns-$vni#d" $vrrp_conf
[ -f "$pid_file" ] && ip netns exec $router kill -HUP $(cat $pid_file)
grep -q " dev ns-$vni" $vrrp_conf
if [ $? -ne 0 ]; then
    ip link del ln-$vni
    apply_vnic -D ln-$vni
    ./clear_link.sh $vni
fi
