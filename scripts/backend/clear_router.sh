#!/bin/bash

cd `dirname $0`
source ../cloudrc

[ $# -lt 2 ] && echo "$0 <router> <vrrp_vni>" && exit -1

router=$1
vrrp_vni=$2
[ -z "$router" ] && exit 1

suffix=${router%%-*}
ip netns exec $router ip link set lo down
ip link del te-$suffix
apply_vnic -D te-$suffix
ip link del ti-$suffix
apply_vnic -D ti-$suffix
./clear_link.sh $vrrp_vni
while read line; do
    [ -z "$line" ] && continue 
    vni=$line
    ip link del ln-$vni
    ./clear_link.sh $vni
done
udevadm settle
ip netns del $router
router_dir=/opt/cloudland/cache/router/$router
kill $(cat $router_dir/keepalived.pid)
rm -rf $router_dir
[ "$RECOVER" = "true" ] || sql_exec "delete from router where name='$router'"
[ "$RECOVER" = "true" ] || sql_exec "delete from gateway where router='$router'"
