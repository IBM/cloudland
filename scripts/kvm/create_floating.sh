#!/bin/bash

cd `dirname $0`
source ../cloudrc

[ $# -lt 6 ] && echo "$0 <router> <ext_ip> <ext_gw> <ext_vlan> <int_ip> <int_vlan>" && exit -1

ID=$1
router=router-$1
ext_cidr=$2
ext_ip=${2%/*}
ext_gw=${3%/*}
ext_vlan=$4
int_addr=$5
int_ip=${int_addr%/*}
int_vlan=$6

[ -z "$router" -o "$router" = "router-0" -o  -z "$ext_ip" -o -z "$int_ip" ] && exit 1
ip netns list | grep -q $router
[ $? -ne 0 ] && echo "Router $router does not exist" && exit -1

rtID=$(( $ext_vlan % 250 + 2 ))
table=fip-$ext_vlan
rt_file=/etc/iproute2/rt_tables
grep "^$rtID $table" $rt_file
if [ $? -ne 0 ]; then
    for i in {1..250}; do
        grep "^$rtID\s" $rt_file
	[ $? -ne 0 ] && break
        rtID=$(( ($rtID + 17) % 250 + 2 ))
    done
    echo "$rtID $table" >>$rt_file
fi
suffix=${ID}-${ext_vlan}
ext_dev=te-$suffix
./create_veth.sh $router ext-$suffix te-$suffix
ip link set te-$suffix netns $router 

ip netns exec $router ip addr add $ext_cidr dev $ext_dev
ip netns exec $router ip route add default via $ext_gw table $table
ip_net=$(ipcalc -b $int_addr | grep Network | awk '{print $2}')
ip netns exec $router ip route add $ip_net dev ns-$int_vlan table $table
ip netns exec $router ip rule add from $int_ip lookup $table
ip netns exec $router ip rule add to $int_ip lookup $table
ip netns exec $router iptables -t nat -I POSTROUTING -s $int_ip -m set ! --match-set nonat dst -j SNAT --to-source $ext_ip
ip netns exec $router iptables -t nat -I PREROUTING -d $ext_ip -j DNAT --to-destination $int_ip
ip netns exec $router arping -c 3 -I $ext_dev -s $ext_ip $ext_ip

router_dir=/opt/cloudland/cache/router/$router
ip netns exec $router iptables-save > $router_dir/iptables.save
