#!/bin/bash

cd `dirname $0`
export RECOVER="true"
export SCI_CLIENT_ID=$(ls /opt/cloudland/db/resource-?.db | sed 's#.*-\(.\).db#\1#')
source ../cloudrc

sql_exec "select * from router" |
while read line; do
    router=$(echo $line | cut -d'|' -f1)
    int_ip=$(echo $line | cut -d'|' -f2)
    ext_ip=$(echo $line | cut -d'|' -f4)
    vrrp_vni=$(echo $line | cut -d'|' -f6)
    vrrp_ip=$(echo $line | cut -d'|' -f7)
    sql_exec "select gateway_ip, subnet_vni from gateway where router='$router'" | tr '|' ' ' | ./create_router.sh $router $ext_ip $int_ip $vrrp_vni $vrrp_ip
done

sql_exec "select router,ext_type,floating_ip,internal_ip from floating" |
while read line; do
    echo $line | tr '|' ' ' | xargs ./create_floating.sh
done
