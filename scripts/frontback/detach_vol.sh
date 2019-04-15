#!/bin/bash

cd `dirname $0`
source ../cloudrc

[ $# -lt 3 ] && echo "$0 <vm_ID> <volume> <device>" && exit -1

vm_ID=$1
vol_name=$2
device=$3

sql_exec "update volume set inst_id='', device='', status='available' where name='$vol_name'"
