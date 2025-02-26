#!/bin/bash

cd `dirname $0`
source ../../cloudrc

[ $# -lt 1 ] && echo "$0 <interface> [force]" && exit -1

vnic=$1
force=$2
chain_in=secgroup-in-$vnic
chain_out=secgroup-out-$vnic
chain_as=secgroup-as-$vnic

if [ "$force" != "true" ]; then 
    for i in {1..35}; do
        [ -f $async_job_dir/$vnic ] && break
        sleep 1
    done
    rm -f $async_job_dir/$vnic
fi

apply_fw -D FORWARD -m physdev --physdev-out $vnic --physdev-is-bridged -j secgroup-chain
apply_fw -D FORWARD -m physdev --physdev-in $vnic --physdev-is-bridged -j secgroup-chain
apply_fw -D secgroup-chain -m physdev --physdev-out $vnic --physdev-is-bridged -j $chain_in
apply_fw -D secgroup-chain -m physdev --physdev-in $vnic --physdev-is-bridged -j $chain_out
apply_fw -D INPUT -m physdev --physdev-in $vnic --physdev-is-bridged -j $chain_out

apply_fw -F $chain_in
apply_fw -F $chain_as
apply_fw -F $chain_out
apply_fw -X $chain_in
apply_fw -X $chain_as
apply_fw -X $chain_out
