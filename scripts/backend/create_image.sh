#!/bin/bash

cd $(dirname $0)
source ../cloudrc

[ $# -lt 2 ] && die "$0 <ID> <url>"

ID=$1
url=$2

state=error
image=$image_cache/image-$1
wget -q $url -O $image
format=$(qemu-img info $image | grep 'file format' | cut -d' ' -f3)
[ "$format" = "qcow2" -o "$format" = "raw" ] && state=available
[ ! -s "$image" ] && state=error
mv $image ${image}.${format}
sync_target /opt/cloudland/cache/image
echo "|:-COMMAND-:| $(basename $0) '$ID' '$state' '$format'"
