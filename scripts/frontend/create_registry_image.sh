#!/bin/bash

cd $(dirname $0)
source ../cloudrc

[ $# -lt 7 ] && die "$0 <ID> <version> <initramfs> <kernel> <image> <installer> <cli>  <virt_type>"

ID=$1
version=$2
initramfs=$3
kernel=$4
image=$5
installer=$6
cli=$7
virt_type=$8
access_addr=$9

base_dir=$image_cache/ocp/$version/$virt_type
mkdir -p $base_dir
#sync_target /opt/cloudland/cache/image
curl $initramfs -o $base_dir/rhcos-installer-initramfs.img
curl $kernel -o $base_dir/rhcos-installer-kernel
curl $image -o $base_dir/rhcos-rootfs.img
curl $installer -o $base_dir/openshift-install-linux.tgz
curl $cli -o $base_dir/openshift-client-linux.tgz


#create ocd.config file
filePath=/opt/cloudland/web/clui/public/misc/openshift/ocd.conf.$version.$virt_type

cat > $filePath << EOF
coreos_rootfs_url=http://$access_addr:4000/ocp/$version/$virt_type/rhcos-rootfs.img
openshift_installer=http://$access_addr:4000/ocp/$version/$virt_type/openshift-install-linux.tgz
openshift_client=http://$access_addr:4000/ocp/$version/$virt_type/openshift-client-linux.tgz
EOF

echo "|:-COMMAND-:| $(basename $0) '$ID' '$base_dir' "
