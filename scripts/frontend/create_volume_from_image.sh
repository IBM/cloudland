#!/bin/bash

cd `dirname $0`
source ../cloudrc

[ $# -lt 2 ] && echo "$0 <user> <image> [disk_inc] [description]" && exit -1

owner=$1
img_name=$2
disk_inc=$3
desc=$4

img_file=$cache_dir/$img_name
[ -f "$img_file" ] || die "Image does not exist!"

vol_name=`date +%m%d%H%M%S-%N`
sql_exec "insert into volume (name, owner, description, bootable, status) values ('$vol_name', '$owner', '$desc', 'true', 'creating')"
/opt/cloudland/bin/sendmsg "inter" "/opt/cloudland/scripts/backend/`basename $0` '$vol_name' '$img_name' '$disk_inc'"
echo "$vol_name|creating"
