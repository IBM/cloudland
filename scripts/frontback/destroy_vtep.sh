#!/bin/bash

cd `dirname $0`
source ../cloudrc

[ $# -lt 1 ] && die "$0 <instance>"

instance=$1

sql_rt=$(sql_exec "select vni, inner_ip from vtep where instance='$instance'")
vni=$(echo $sql_rt | cut -d'|' -f1)
ip=$(echo $sql_rt | cut -d'|' -f2)
vpc=$(sql_exec "select vpc_name from netlink where vlan=$vni")
group=vx$vni
sql_exec "update vtep set outer_ip='', status='inactive' where instance='$instance'"
curl -H "Content-Type: application/json" "$FLIGHT_PLANNER_ENDPOINT/routes/$vpc/$ip" -XDELETE"
echo "$instance|destroyed"
