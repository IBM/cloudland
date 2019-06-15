#!/bin/bash

cd `dirname $0`
source ../cloudrc

[ $# -lt 4 ] && echo "$0 <router> <ext_type> <ext_ip> <int_ip>" && exit -1

router=router-$1
ext_type=$2
ext_ip=${3%/*}
int_ip=${4%/*}

[ -z "$router" -o -z "$ext_ip" -o -z "$int_ip" ] && exit 1

if [ "$ext_type" = "public" ]; then
    ext_dev=$(ip netns exec $router ip -o link show | cut -d: -f2 | cut -d@ -f1 | grep ext | xargs)
    ip netns exec $router iptables -t nat -D PREROUTING -d $ext_ip -i $ext_dev -j DNAT --to-destination $int_ip
    ip netns exec $router iptables -t nat -D POSTROUTING -s $int_ip ! -d 10.0.0.0/8 -o $ext_dev -j SNAT --to-source $ext_ip
elif [ "$ext_type" = "private" ]; then
    ext_dev=$(ip netns exec $router ip -o link show | cut -d: -f2 | cut -d@ -f1 | grep int | xargs)
    ip netns exec $router iptables -t nat -D PREROUTING -d $ext_ip -i $ext_dev -j DNAT --to-destination $int_ip
    ip netns exec $router iptables -t nat -D POSTROUTING -s $int_ip -d 10.0.0.0/8 -o $ext_dev -j SNAT --to-source $ext_ip
fi

router_dir=/opt/cloudland/cache/router/$router
vrrp_conf=$router_dir/keepalived.conf
notify_sh=$router_dir/notify.sh
pid_file=$router_dir/keepalived.pid
sed -i "\#$ext_ip/32 dev $ext_dev#d" $vrrp_conf
sed -i "\#ip netns exec $router arping -c 1 -S $ext_ip $ext_gw#d" $notify_sh
[ -f "$pid_file" ] && ip netns exec $router kill -HUP $(cat $pid_file)
ip netns exec $router iptables-save > $router_dir/iptables.save
