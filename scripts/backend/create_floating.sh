#!/bin/bash

cd `dirname $0`
source ../cloudrc

[ $# -lt 4 ] && echo "$0 <router> <ext_type> <ext_ip> <int_ip>" && exit -1

router=router-$1
ext_type=$2
ext_ip=$3
int_ip=$4

[ -z "$router" -o  -z "$ext_ip" -o -z "$int_ip" ] && exit 1
ip netns list | grep -q $router
[ $? -ne 0 ] && exit 0

if [ "$ext_type" = "public" ]; then
    ext_dev=$(ip netns exec $router ip -o link show | cut -d: -f2 | cut -d@ -f1 | grep ext | xargs)
    ext_gw=$(ip netns exec $router ip route | grep default | cut -d' ' -f3)
    ip netns exec $router iptables -t nat -I PREROUTING -d $ext_ip -i $ext_dev -j DNAT --to-destination $int_ip
    ip netns exec $router iptables -t nat -I POSTROUTING -s $int_ip ! -d 10.0.0.0/8 -o $ext_dev -j SNAT --to-source $ext_ip
elif [ "$ext_type" = "private" ]; then
    ext_dev=$(ip netns exec $router ip -o link show | cut -d: -f2 | cut -d@ -f1 | grep int | xargs)
    ext_gw=$(ip netns exec $router ip route | grep 10.0.0.0 | cut -d' ' -f3)
    ip netns exec $router iptables -t nat -I PREROUTING -d $ext_ip -i $ext_dev -j DNAT --to-destination $int_ip
    ip netns exec $router iptables -t nat -I POSTROUTING -s $int_ip -d 10.0.0.0/8 -o $ext_dev -j SNAT --to-source $ext_ip
fi
ip netns exec $router arping -c 2 -S $ext_ip $ext_ip

router_dir=/opt/cloudland/cache/router/$router
vrrp_conf=$router_dir/keepalived.conf
notify_sh=$router_dir/notify.sh
pid_file=$router_dir/keepalived.pid
sed -i "\#$ext_ip/32 dev $ext_dev#d" $vrrp_conf
sed -i "/virtual_ipaddress {/a $ext_ip/32 dev $ext_dev" $vrrp_conf
sed -i "\#ip netns exec $router arping -c 1 -S $ext_ip $ext_gw#d" $notify_sh
sed -i "/\"MASTER\")/a ip netns exec $router arping -c 1 -S $ext_ip $ext_gw" $notify_sh
[ -f "$pid_file" ] && ip netns exec $router kill -HUP $(cat $pid_file)
[ "$RECOVER" = "true" ] || sql_exec "insert into floating values ('$router', '$ext_type', '$ext_dev', '$ext_ip', '$int_ip')"
