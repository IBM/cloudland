#!/bin/bash

cd $(dirname $0)
source ../cloudrc

[ $# -lt 2 ] && echo "$0 <vol_ID> <size>" && exit -1

vol_ID=$1
size=$2
state='error'

qemu-img create -f qcow2 -o cluster_size=2M $volume_dir/volume-${vol_ID}.disk ${size}G
[ $? -eq 0 ] && state='available'
echo "|:-COMMAND-:| $(basename $0) '$vol_ID' 'volume-${vol_ID}.disk' '$state'"
