#!/bin/bash

cd `dirname $0`
source ../cloudrc

[ $# -lt 4 ] && echo "$0 <router> <ext_type> <ext_ip> <int_ip>" && exit -1

ID=$1
router=router-$1
ext_type=$2
ext_ip=${3%/*}
int_ip=${4%/*}
int_net=$4

[ -z "$router" -o  -z "$ext_ip" -o -z "$int_ip" ] && exit 1
ip netns list | grep -q $router
[ $? -ne 0 ] && exit 0

router_dir=/opt/cloudland/cache/router/$router
vrrp_conf=$router_dir/keepalived.conf
notify_sh=$router_dir/notify.sh
pid_file=$router_dir/keepalived.pid
dest_nets=$(grep 'add -net' $notify_sh | awk '{print $8}')

if [ "$ext_type" = "public" ]; then
    ext_dev=te-$ID
    ip netns exec $router iptables -t nat -I POSTROUTING -s $int_ip -m set ! --match-set nonat dst -j SNAT --to-source $ext_ip
elif [ "$ext_type" = "private" ]; then
    ext_dev=ti-$ID
else
    echo "Rong routing type" && exit 1
fi
ip netns exec $router iptables -t nat -I PREROUTING -d $ext_ip -j DNAT --to-destination $int_ip
ip netns exec $router iptables -t mangle -A PREROUTING -s $int_net -d $ext_ip -j MARK --set-xmark 0x400
ip netns exec $router arping -c 3 -I $ext_dev -s $ext_ip $ext_ip

sed -i "\#$ext_ip/32 dev $ext_dev#d" $vrrp_conf
sed -i "/virtual_ipaddress {/a $ext_ip/32 dev $ext_dev" $vrrp_conf
sed -i "\#ip netns exec $router arping -c . -I $ext_dev -s $ext_ip $ext_ip#d" $notify_sh
sed -i "/\"MASTER\")/a ip netns exec $router arping -c 3 -I $ext_dev -s $ext_ip $ext_ip" $notify_sh
[ -f "$pid_file" ] && ip netns exec $router kill -HUP $(cat $pid_file)
ip netns exec $router iptables-save > $router_dir/iptables.save
