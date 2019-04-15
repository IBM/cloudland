#!/bin/bash

cd `dirname $0`
source ../cloudrc

[ $# -lt 2 ] && echo "$0 <user> <url> [shared(true|false)] [description] [platform]" && exit -1

owner=$1
url=$2
shared=$3
desc=$4
platform=$5

fname=`basename $url`
[ -f "$cache_dir/$fname" -o -f "$volume_dir/$fname" ] && die "Image with the same name exists, please rename it first!"

/opt/cloudland/bin/sendmsg "exec" "/opt/cloudland/scripts/frontend/download_upload.sh '$owner' '$url' '$shared' '$desc' '$platform'"
echo "$fname|uploading"
