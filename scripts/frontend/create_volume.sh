#!/bin/bash

cd `dirname $0`
source ../cloudrc

[ $# -lt 2 ] && echo "$0 <user> <size> [description]" && exit -1

owner=$1
vol_size=$2
desc=$3

vol_size=${vol_size%%[gG]}
[ $vol_size -gt 0 -a $vol_size -le $vol_limit ] || die "Invalid volume size!"
vol_name=`date +%m%d%H%M%S-%N`
vol_file=$volume_dir/$vol_name.disk
dd if=/dev/zero of=$vol_file count=0 bs=1073741824 seek=$vol_size >/dev/null 2>&1
sql_exec "insert into volume (name, size, owner, description, status) values ('$vol_name', '$vol_size', '$owner', '$desc', 'available')"
echo "$vol_name|created"
