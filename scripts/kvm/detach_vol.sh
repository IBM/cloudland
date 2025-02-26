#!/bin/bash

cd `dirname $0`
source ../cloudrc

[ $# -lt 3 ] && echo "$0 <vm_ID> <volume> <device>" && exit -1

vm_ID=$1
vol_name=$2
device=$3
vol_file=$volume_dir/$vol_name.disk

virsh detach-disk $vm_ID $device
[ $? -eq 0 ] && echo "|:-COMMAND-:| /opt/cloudland/scripts/frontback/`basename $0` '$vm_ID' '$vol_name' '$device'"
