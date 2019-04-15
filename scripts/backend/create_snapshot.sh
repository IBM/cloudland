#!/bin/bash

cd `dirname $0`
source ../cloudrc

[ $# -lt 1 ] && echo "$0 <vm_ID>" && exit -1

vm_ID=$1
snap_file=$snapshot_dir/$vm_ID.qcow2
stat='failed'

virsh suspend $vm_ID
image=${image_dir}/$vm_ID.qcow2

qemu-img convert -f qcow2 -O qcow2 $image $snap_file
snap_size=`ls -l $snap_file | cut -d' ' -f5`
[ $? -eq 0 ] && stat='created'
virsh resume $vm_ID
echo "|:-COMMAND-:| /opt/cloudland/scripts/frontback/`basename $0` '$vm_ID' '$snap_size' '$stat'"
