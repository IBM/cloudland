#!/bin/bash

cd `dirname $0`
source ../cloudrc

[ $# -lt 2 ] && echo "$0 <router> <vni> [type]" && exit -1

router=$1
[ "${router/router-/}" = "$router" ] && router=router-$1
ID=${router/router-/}
vni=$2
rtype=$3

routes=$(cat)
router_dir=/opt/cloudland/cache/router/$router
vrrp_conf=$router_dir/keepalived.conf
notify_sh=$router_dir/notify.sh
pid_file=$router_dir/keepalived.pid

iface=ns-$vni
[ "$rtype" = "private" ] && iface=ti-$ID
[ "$rtype" = "pubilc" ] && iface=te-$ID
sed -i "\#ip netns exec $router route add -net .* gw .* dev $iface#d" $notify_sh
i=0
rlen=$(jq length <<< $routes)
while [ $i -lt $rlen ]; do
    destination=$(jq -r .[$i].destination <<< $routes)
    nexthop=$(jq -r .[$i].nexthop <<< $routes)
    if [ "$rtype" != "public" ]; then
        ip netns exec $router ipset add nonat $destination
    fi
    if [ "$rtype" = "private" ]; then
        ip netns exec $router iptables -t nat -A POSTROUTING -d $destination -j SNAT --to-source $nexthop
    fi
    ip netns exec $router route add -net $destination gw $nexthop dev $iface
    sed -i "/\"MASTER\")/a ip netns exec $router route add -net $destination gw $nexthop dev $iface" $notify_sh
    let i=$i+1
done

[ -f "$pid_file" ] && ip netns exec $router kill -HUP $(cat $pid_file)
ip netns exec $router iptables-save > $router_dir/iptables.save
ip netns exec $router ipset save > $router_dir/ipset.save
