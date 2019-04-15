#!/bin/bash

cd `dirname $0`
source ../cloudrc

[ $# -lt 1 ] && echo "$0 <user> [vm_ID]" && exit -1

owner=$1
vm_ID=$2
if [ -n "$vm_ID" ]; then
    sql_ret=`sql_exec "select inst_id, image, ip_addr, hname, vlan, status, vnc from instance where owner='$owner' and inst_id='$vm_ID' and status!='deleted'"`
else
    sql_ret=`sql_exec "select inst_id, image, ip_addr, hname, vlan, status, vnc from instance where owner='$owner' and status!='deleted'"`
fi

echo "$sql_ret"
