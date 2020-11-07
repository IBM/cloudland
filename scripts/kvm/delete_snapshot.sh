#!/bin/bash

cd `dirname $0`
source ../cloudrc

[ $# -lt 1 ] && echo "$0 <vm_ID>" && exit -1

vm_ID=$1
snap_file=$snapshot_dir/$vm_ID.qcow2

rm -f $snap_file
