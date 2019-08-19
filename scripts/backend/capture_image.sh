#!/bin/bash

cd `dirname $0`
source ../cloudrc

[ $# -lt 2 ] && echo "$0 <img_ID> <vm_ID>" && exit -1

img_ID=$1
vm_ID=inst-$2
state=error
image=$image_cache/image-$1

virsh suspend $vm_ID
image=${image_dir}/image-$vm_ID.qcow2
inst_img=$cache_dir/instance/${vm_ID}.disk

format=$(qemu-img info $inst_img | grep 'file format' | cut -d' ' -f3)
qemu-img convert -f $format -O qcow2 $inst_img $image
[ -s "$image" ] && state=available
virsh resume $vm_ID
sync_target /opt/cloudland/cache/image/
echo "|:-COMMAND-:| create_image.sh '$img_ID' '$state' 'qcow2'"
