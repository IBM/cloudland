#!/bin/bash

cd `dirname $0`
source ../cloudrc

[ $# -lt 2 ] && echo "$0 <router> <vlan> [gateways]" && exit -1

router=$1
vlan=$2
shift 2
gateways=$@

[ "$vlan" == "$external_vlan" ] && die "Can not attach external vlan"

tap_dev=rtap-$vlan
ns_dev=rns-$vlan
ip link add $ns_dev type veth peer name $tap_dev
./create_link.sh $vlan
ip link set dev $tap_dev master br$vlan
ip link set $tap_dev up
ip link set $ns_dev netns $router
ip netns exec $router ip link set $ns_dev up

for gw in $gateways; do
    gw_ip=`echo $gw | cut -d'|' -f1`
    gw_mask=`echo $gw | cut -d'|' -f2`
    ./set_gateway.sh $router $vlan $gw_ip $gw_mask
done
echo "|:-COMMAND-:| /opt/cloudland/scripts/frontback/`basename $0` $router $vlan"
