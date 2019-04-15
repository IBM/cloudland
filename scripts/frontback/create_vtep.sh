#!/bin/bash -xv

cd `dirname $0`
source ../cloudrc

[ $# -lt 3 ] && die "$0 <instance> <hyper_id> <hyper_ip>"

instance=$1
hyper_id=$2
hyper_ip=$3

sql_rt=$(sql_exec "select vni, inner_ip, inner_mac from vtep where instance='$instance'")
vni=$(echo $sql_rt | cut -d'|' -f1)
ip=$(echo $sql_rt | cut -d'|' -f2)
mac=$(echo $sql_rt | cut -d'|' -f3)
vpc=$(sql_exec "select vpc_name from netlink where vlan=$vni")
network=$(sql_exec "select network from address where IP='$ip'")
ip_prefix=172.16.11.0/24
json=$(jq -n -r --arg vtep_ip ${hyper_ip} --arg vtep_port 4567 --arg dst_mac $mac\
       --arg subnet_prefix "$ip_prefix" --arg vm_interface_name "${instance}-vs"\
       '{subnet_prefix: $subnet_prefix, vtep_ip: $vtep_ip, vtep_port: $vtep_port, dst_mac: $dst_mac, vm_interface_name: $vm_interface_name}')
found=false
members=$(sql_exec "select distinct(hyper_id) from vtep where vni=$vni and status='active'")
for i in $members; do
    if [ $hyper_id -eq $i ]; then
        found=true
        break
    fi
done
group=vx$vni
if [ "$found" = "false" ]; then
    grp_members=$(echo $members $hyper_id | tr ' ' ',')
    /opt/cloudland/bin/grpcmsg "0" "mkgrp=$group:$grp_members"
    sql_rt=$(sql_exec "select instance, inner_ip, inner_mac, outer_ip from vtep where vni=$vni and hyper_id!=$hyper_id and status='active'")
    echo "$sql_rt" | while read line; do
        rinst=$(echo $line | cut -d'|' -f1)
        rip=$(echo $line | cut -d'|' -f2)
        rmac=$(echo $line | cut -d'|' -f3)
        rhyper=$(echo $line | cut -d'|' -f4)
        /opt/cloudland/bin/grpcmsg "0" "inter=$hyper_id" "/opt/cloudland/scripts/backend/add_fdb.sh '$rinst' $vni '$rip' '192.168.1.1' '$rmac'"
    done
fi
curl -H "Content-Type: application/json" "$FLIGHT_PLANNER_ENDPOINT/routes/$vpc/$ip" -XPUT -d "$json"
sql_exec "update vtep set outer_ip='$hyper_ip', hyper_id=$hyper_id, status='active' where instance='$instance'"
echo "$instance|created"
