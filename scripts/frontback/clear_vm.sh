#!/bin/bash

cd `dirname $0`
source ../cloudrc

[ $# -lt 1 ] && echo "$0 <vm_ID>" && exit -1

vm_ID=$1

sql_exec "update volume set inst_id='', device='' where inst_id='$vm_ID'"
ids=`sql_exec "select id from address where instance='$vm_ID' and allocated='true'"`
for i in $ids; do
    sql_ret=`sql_exec "select vlan, IP, mac from address where id='$i'"`
    vl=`echo $sql_ret | cut -d'|' -f1`
    ip=`echo $sql_ret | cut -d'|' -f2`
    mac=`echo $sql_ret | cut -d'|' -f3`
    rt=`sql_exec "select id from compute where hyper_name=(select dh_host from netlink where vlan='$vl')"`
    /opt/cloudland/bin/sendmsg "inter $rt" "/opt/cloudland/scripts/backend/del_host.sh $vl $mac $ip"
    sql_exec "update address set allocated='false' where IP='$ip'"
done

