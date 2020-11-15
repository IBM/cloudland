#!/bin/bash

cd `dirname $0`
source ../cloudrc

[ $# -lt 2 ] && echo "$0 <volume> <image> [disk_inc]" && exit -1

vol_name=$1
img_name=$2
disk_inc=$3
vol_stat=error

img_file=$cache_dir/$img_name
[ -f "$img_file" ] || die "Image does not exist!"

vol_file=$volume_dir/$vol_name.disk
qemu-img convert -f qcow2 -O raw $img_file $vol_file
[ $? -eq 0 ] && vol_stat=available
if [ -n "$disk_inc" ]; then
    qemu-img resize $vol_file +${disk_inc%%[G|g]}G
fi
vol_size=`qemu-img info $vol_file | grep 'virtual size:' | cut -d' ' -f3`
echo "|:-COMMAND-:| /opt/cloudland/scripts/frontback/`basename $0` '$vol_name' '$vol_size' '$vol_stat'"
