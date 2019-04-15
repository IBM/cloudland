#!/bin/bash

cd `dirname $0`
source ../cloudrc

[ $# -lt 3 ] && echo "$0 <user> <vm_ID> <volume>" && exit -1

owner=$1
vm_ID=$2
vol_name=$3

num=`sql_exec "select count(*) from volume where name='$vol_name' and owner='$owner' and coalesce(inst_id, '')=''"`
[ $num -eq 1 ] || die "Wrong volume status or not volume owner!"
num=`sql_exec "select count(*) from instance where inst_id='$vm_ID' and owner='$owner' and status='running'"`
[ $num -eq 1 ] || die "VM not running or not VM owner!"
num=`sql_exec "select count(*) from volume where inst_id='$vm_ID'"`
[ $num -le 20 ] || die "Too many volumes attached to VM $vm_ID!"

sql_exec "update volume set inst_id='$vm_ID', where name='$vol_name'"
letters="bcdefghijklmnopqrstuvwxyz"
let num=$num+1
lett=`echo $letters | cut -c $num`
device=vd"$lett"

hyper_id=`sql_exec "select id from compute where hyper_name=(select hyper_name from instance where inst_id='$vm_ID')"`
[ -n "$hyper_id" ] && /opt/cloudland/bin/sendmsg "inter $hyper_id" "/opt/cloudland/scripts/backend/`basename $0` $vm_ID $vol_name $device"
echo "$vm_ID|$vol_name|attaching"
