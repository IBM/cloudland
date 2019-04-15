#!/bin/bash

cd `dirname $0`
source ../cloudrc

[ $# -lt 1 ] && die "$0 <instance>"

instance=$1

ovs-vsctl del-port ${instance}-vs
ip link del ${instance}-vs
ip netns exec $instance ip link set lo down
ip netns del $instance
echo "|:-COMMAND-:| /opt/cloudland/scripts/frontback/$(basename $0) '$instance'"
