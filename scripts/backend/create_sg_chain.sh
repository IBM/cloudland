#!/bin/bash

cd `dirname $0`
source ../cloudrc

[ $# -lt 3 ] && echo "$0 <interface> <ip> <mac>" && exit -1

vnic=$1
ip=${2%%/*}
mac=$3

apply_fw -I FORWARD -m physdev --physdev-out $vnic --physdev-is-bridged -j secgroup-chain
apply_fw -I FORWARD -m physdev --physdev-in $vnic --physdev-is-bridged -j secgroup-chain

chain_in=secgroup-in-$vnic
apply_fw -N $chain_in
apply_fw -I secgroup-chain -m physdev --physdev-out $vnic --physdev-is-bridged -j $chain_in
apply_fw -A $chain_in -m state --state RELATED,ESTABLISHED -j RETURN
apply_fw -A $chain_in -m state --state INVALID -j DROP
apply_fw -A $chain_in -j DROP

chain_out=secgroup-out-$vnic
chain_as=secgroup-as-$vnic
apply_fw -N $chain_as
apply_fw -A $chain_as -s $ip/32 -m mac --mac-source $mac -j RETURN
apply_fw -A $chain_as -j DROP

apply_fw -N $chain_out
apply_fw -I secgroup-chain -m physdev --physdev-in $vnic --physdev-is-bridged -j $chain_out
apply_fw -I INPUT -m physdev --physdev-in $vnic --physdev-is-bridged -j $chain_out
apply_fw -A $chain_out -j $chain_as
apply_fw -A $chain_out -m state --state RELATED,ESTABLISHED -j RETURN
apply_fw -A $chain_out -m state --state INVALID -j DROP
apply_fw -A $chain_out -j DROP

service iptables save
