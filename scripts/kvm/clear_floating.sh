#!/bin/bash

cd `dirname $0`
source ../cloudrc

[ $# -lt 5 ] && echo "$0 <router> <ext_ip> <int_ip> <int_vlan> <mark_id>" && exit -1

ID=$1
router=router-$1
ext_addr=$2
ext_ip=${2%/*}
int_ip=${3%/*}
int_vlan=$4
mark_id=$(($5 % 4294967295))

[ -z "$router" -o -z "$ext_ip" -o -z "$int_ip" ] && exit 1

ext_dev=$(ip netns exec $router ip -o addr | grep "$ext_addr" | awk '{print $2}')
table=fip-${ext_dev##*-}
ip netns exec $router ip rule del from $int_ip lookup $table
ip netns exec $router ip rule del to $int_ip lookup $table
ip netns exec $router ip addr del $ext_addr dev $ext_dev
#ip netns exec $router iptables -t nat -D POSTROUTING -s $int_ip -m set ! --match-set nonat dst -j SNAT --to-source $ext_ip
ip netns exec $router iptables -t nat -D PREROUTING -d $ext_ip -j DNAT --to-destination $int_ip
ip netns exec $router iptables -t nat -D POSTROUTING -s $int_ip -j SNAT --to-source $ext_ip
ip netns exec $router ip addr show $ext_dev | grep 'inet '
if [ $? -ne 0 ]; then
    ip netns exec $router ip link del $ext_dev
fi
ip netns exec $router iptables -t mangle -D PREROUTING -d $ext_ip -j MARK --set-mark $mark_id
ip netns exec $router iptables -S | grep "mark $(printf "0x%x" $mark_id)" | while read line; do
    echo $line | cut -d' ' -f2- | xargs ip netns exec $router iptables -D
done
ip netns exec $router tc filter del dev ns-$int_vlan protocol ip parent 1:0 prio $mark_id handle $mark_id fw flowid 1:$mark_id
ip netns exec $router tc class del dev ns-$int_vlan parent 1: classid 1:$mark_id
ip netns exec $router tc filter del dev $ext_dev protocol ip parent 1:0 prio $mark_id u32 match ip src $ext_ip/32 flowid 1:$mark_id
ip netns exec $router tc class del dev $ext_dev parent 1: classid 1:$mark_id
exit 0
