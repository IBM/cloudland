#!/bin/bash

cd `dirname $0`
source ../cloudrc

[ $# -lt 3 ] && echo "$0 <vm_ID> <size> <status>" && exit -1

vm_ID=$1
snap_size=$2
stat=$3

[ -z "$snap_size" ] && snap_size=0
sql_exec "update snapshot set size='$snap_size', status='$stat' where inst_id='$vm_ID'"
