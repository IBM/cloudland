#!/bin/bash

cd $(dirname $0)
source ../cloudrc

[ $# -lt 8 ] && die "$0 <ID> <version> <initramfs> <kernel> <image> <installer> <cli> <kubelet> <virt_type>"

ID=$1
version=$2
initramfs=$3
kernel=$4
image=$5
installer=$6
cli=$7
kubelet=$8
virt_type=$9

base_dir=$image_cache/$version/$virt_type
mkdir -p $base_dir

wget $initramfs -o $base_dir/rhcos-installer-initramfs.img
wget $kernel -o $base_dir/rhcos-installer-kernel
wget $image -o $base_dir/image
wget $installer -o $base_dir/install
wget $cli -o $base_dir/cli
wget $kubelet -o $base_dir/kubelet
#sync_target /opt/cloudland/cache/image
echo "|:-COMMAND-:| $(basename $0) '$ID' '$base_dir' "
