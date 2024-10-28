#!/bin/bash

cd `dirname $0`
source ../cloudrc

[ $# -lt 4 ] && echo "$0 <router> <ext_ip> <int_ip> <pub_gw>" && exit -1

ID=$1
router=router-$1
ext_cidr=$2
ext_ip=${2%/*}
int_ip=${3%/*}
pub_gw=${4%/*}

[ -z "$router" -o  -z "$ext_ip" -o -z "$int_ip" ] && exit 1
ip netns list | grep -q $router
[ $? -ne 0 ] && echo "Router $router does not exist" && exit -1

ext_dev=te-$ID
ip netns exec $router ip addr add $ext_cidr dev $ext_dev
ip netns exec $router ip route add default via $pub_gw table fip
ip netns exec $router ip rule add from $int_ip lookup fip
ip netns exec $router ip rule add to $int_ip lookup fip
ip netns exec $router iptables -t nat -I POSTROUTING -s $int_ip -m set ! --match-set nonat dst -j SNAT --to-source $ext_ip
ip netns exec $router iptables -t nat -I PREROUTING -d $ext_ip -j DNAT --to-destination $int_ip
ip netns exec $router arping -c 3 -I $ext_dev -s $ext_ip $ext_ip

router_dir=/opt/cloudland/cache/router/$router
ip netns exec $router iptables-save > $router_dir/iptables.save
