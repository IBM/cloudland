#!/bin/bash

cd `dirname $0`
source ../cloudrc

[ $# -lt 1 ] && echo "$0 <name_space>" && exit -1

nspace=$1 

dev_list=$(ip netns exec $nspace  cat /proc/net/dev | tail -n +3 | cut -d: -f1)
if [ ${nspace} != ${nspace##vlan} ]; then
    for dev in $dev_list; do
        link_no=${dev##ns-}
        if [ -n "$link_no" ]; then
            ip link del tap-$link_no
            apply_vnic -D tap-$link_no
        fi
    done
    rm -rf $dmasq_dir/$nspace
elif [ ${nspace} != ${nspace##router-} ]; then
    for dev in $dev_list; do
        link_no=${dev##ns-}
        if [ -n "$link_no" -a "$link_no" != "$dev" ]; then
            ip link del ln-$link_no
            apply_vnic -D ln-$link_no
        fi
        ext_no=${dev##te-}
        if [ -n "$ext_no" -a "$ext_no" != "$dev" ]; then
            ip link del ext-$ext_no
            apply_vnic -D ext-$ext_no
        fi
        int_no=${dev##ti-}
        if [ -n "$int_no" -a "$int_no" != "$dev" ]; then
            ip link del int-$int_no
            apply_vnic -D int-$ext_no
        fi
    done
    rm -rf /opt/cloudland/cache/router/$nspace
fi
ip netns exec $nspace ip link set lo down
ip netns del $nspace
