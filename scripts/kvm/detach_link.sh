#!/bin/bash

cd `dirname $0`
source ../cloudrc

[ $# -lt 3 ] && echo "$0 <router> <vlan> <gateways>" && exit -1

router=$1
vlan=$2
gateways=$3
tap_dev=rtap-$vlan
for gw in $gateways; do
    gw_ip=`echo $gw | cut -d'|' -f1`
    gw_mask=`echo $gw | cut -d'|' -f2`
    ./clear_gateway.sh $router $vlan $gw_ip $gw_mask
done
ip link del $tap_dev
echo "|:-COMMAND-:| /opt/cloudland/scripts/frontback/`basename $0` $router $vlan"
