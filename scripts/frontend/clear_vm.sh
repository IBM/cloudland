#!/bin/bash

cd `dirname $0`
source /opt/cloudland/scripts/cloudrc

[ $# -lt 2 ] && echo "$0 <user> <vm_ID>" && exit -1

owner=$1
vm_ID=$2
sql_ret=`sql_exec "select vlan, ip_addr, mac_addr, hyper_name from instance where inst_id='$vm_ID' and owner='$owner' and status!='launching'"`
[ -z "$sql_ret" ] && die "No such VM!"

vlan=`echo $sql_ret | cut -d'|' -f1`
vm_ip=`echo $sql_ret | cut -d'|' -f2`
vm_mac=`echo $sql_ret | cut -d'|' -f3`
hyper=`echo $sql_ret | cut -d'|' -f4`

sql_exec "update instance set deleted=datetime('now'), status='deleted' where inst_id='$vm_ID'"
hyper_id=`sql_exec "select id from compute where hyper_name='$hyper'"`
if [ -n "$hyper_id" ]; then
    /opt/cloudland/bin/sendmsg "inter $hyper_id" "/opt/cloudland/scripts/backend/`basename $0` $vm_ID"
else
    /opt/cloudland/bin/sendmsg "toall" "/opt/cloudland/scripts/backend/`basename $0` $vm_ID"
fi

echo "$vm_ID|deleted"
