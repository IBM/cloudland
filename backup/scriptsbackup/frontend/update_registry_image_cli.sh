#!/bin/bash

cd $(dirname $0)
source ../cloudrc

[ $# -lt 4 ] && die "$0 <ID> <version> <cli> <virt_type>"

ID=$1
version=$2
cli=$3
virt_type=$4

base_dir=$image_cache/ocp/$version/$virt_type

#sync_target /opt/cloudland/cache/image
curl $cli -o $base_dir/openshift-client-linux.tgz

echo "|:-COMMAND-:| $(basename $0) '$ID' '$base_dir' "
