#!/bin/bash

cd `dirname $0`
source ../cloudrc

[ $# -lt 2 ] && echo "$0 <user> <url> [shared(true|false)] [description] [platform]" && exit -1

owner=$1
url=$2
shared=$3
desc=$4
platform=$5

file=$cache_dir/`basename $url`
[ -f "$file" ] && die "Image already exists, delete it first!"
/usr/bin/wget -q $url -O $file
[ -f "$file" ] || die "Failed to download file!"

sync
size=`ls -l $file | cut -d' ' -f 5`
if [ $size -eq 0 ]; then
    rm -f $file
    die "Invalid File!"
fi

./upload_image.sh "$owner" "$file" "$shared" "$desc" "$platform"
