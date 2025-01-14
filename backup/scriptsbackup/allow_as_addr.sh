#!/bin/bash

cd `dirname $0`
source ../cloudrc

[ $# -lt 2 ] && echo "$0 <ip> <mac>" && exit -1

vm_ip=${1%%/*}
vm_mac=$2
vnic=tap$(echo $vm_mac | cut -d: -f4- | tr -d :)

chain_as=secgroup-as-$vnic
apply_fw -F $chain_as

apply_fw -I $chain_as -s $vm_ip -m mac --mac-source $vm_mac -j RETURN
pairs=$(cat)
for ap in $pairs; do
    ip=${ap%%-*}
    [ -z "$ip" ] && continue
    mac=${ap##*-}
    [ -z "$mac" -o "$mac" = "$ip" ] && mac=$vm_mac
    apply_fw -I $chain_as -s $ip -m mac --mac-source $mac -j RETURN
done
