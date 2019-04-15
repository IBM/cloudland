#!/bin/bash

cd `dirname $0`
source ../cloudrc

[ $# -lt 5 ] && echo "$0 <vm_ID> <status> <hyper> <mac> <vnc> [size] [is_vol]" && exit -1

vm_ID=$1
vm_stat=$2
hyper=$3
vm_mac=$4
vm_vnc=$5
vsize=$6
is_vol=$7

hyper_ip=`sql_exec "select ip_addr from compute where hyper_name='$hyper'"`
vm_vnc=$hyper_ip:$vm_vnc
sql_exec "update instance set inst_id='$vm_ID', status='$vm_stat', hyper_name='$hyper', vnc='$vm_vnc', active=datetime('now') where mac_addr='$vm_mac'"
[ "$is_vol" == "true" ] && ./attach_vol.sh "$vm_ID" "$vm_ID" "vda" "$vsize"
