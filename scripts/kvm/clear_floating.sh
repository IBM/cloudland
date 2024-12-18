#!/bin/bash

cd `dirname $0`
source ../cloudrc

[ $# -lt 3 ] && echo "$0 <router> <ext_ip> <int_ip>" && exit -1

ID=$1
router=router-$1
ext_addr=$2
ext_ip=${2%/*}
int_ip=${3%/*}

[ -z "$router" -o -z "$ext_ip" -o -z "$int_ip" ] && exit 1

ext_dev=$(ip netns exec $router ip -o addr | grep "$ext_addr" | awk '{print $2}')
table=fip-${ext_dev##*-}
ip netns exec $router ip rule del from $int_ip lookup $table
ip netns exec $router ip rule del to $int_ip lookup $table
ip netns exec $router ip addr del $ext_addr dev $ext_dev
ip netns exec $router iptables -t nat -D POSTROUTING -s $int_ip -m set ! --match-set nonat dst -j SNAT --to-source $ext_ip
ip netns exec $router iptables -t nat -D PREROUTING -d $ext_ip -j DNAT --to-destination $int_ip
