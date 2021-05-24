#!/bin/bash

cd $(dirname $0)
source ../cloudrc

[ $# -lt 4 ] && die "$0 <ID> <version> <image> <virt_type>"

ID=$1
version=$2
image=$3
virt_type=$4

base_dir=$image_cache/ocp/$version/$virt_type

#sync_target /opt/cloudland/cache/image
curl $image -o $base_dir/rhcos-rootfs.img

echo "|:-COMMAND-:| $(basename $0) '$ID' '$base_dir' "
