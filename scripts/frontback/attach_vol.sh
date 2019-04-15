#!/bin/bash

cd `dirname $0`
source ../cloudrc

[ $# -lt 3 ] && echo "$0 <vm_ID> <volume> <device> [size]" && exit -1

vm_ID=$1
vol_name=$2
device=$3
vsize=$4

if [ -z "$vsize" ]; then
    sql_exec "update volume set inst_id='$vm_ID', device='$device' where name='$vol_name'"
else
    vsize=${vsize%%[G|g]}
    sql_exec "update volume set inst_id='$vm_ID', device='$device', size='$vsize', status='in_use' where name='$vol_name'"
fi
