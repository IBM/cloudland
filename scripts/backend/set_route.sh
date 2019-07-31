#!/bin/bash

cd `dirname $0`
source ../cloudrc

[ $# -lt 1 ] && echo "$0 <router>" && exit -1

router=$1
[ "${router/router-/}" = "$router" ] && router=router-$1

routes=$(cat)
router_dir=/opt/cloudland/cache/router/$router
vrrp_conf=$router_dir/keepalived.conf
notify_sh=$router_dir/notify.sh
pid_file=$router_dir/keepalived.pid

i=0
rlen=$(jq length <<< $routes)
while [ $i -lt $rlen ]; do
    destination=$(jq -r .[$i].destination <<< $routes)
    nexthop=$(jq -r .[$i]nexthop <<< $routes)
    sed -i "\#ip netns exec $router route add -net $destination gw $nexthop#d" $notify_sh
    sed -i "/\"MASTER\")/a ip netns exec $router route add -net $destination gw $nexthop" $notify_sh
    let i=$i+1
done

[ -f "$pid_file" ] && ip netns exec $router kill -HUP $(cat $pid_file)
