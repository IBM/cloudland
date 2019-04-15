#!/bin/bash

cd `dirname $0`
source ../cloudrc

[ $# -lt 2 ] && echo "$0 <user> <volume>" && exit -1

owner=$1
vol_name=$2
num=`sql_exec "select count(*) from volume where name='$vol_name' and owner='$owner' and coalesce(inst_id, '')=''"`
[ $num -eq 1 ] || die "Volume not exist or not owner or attached to an instance!"
vol_file=$volume_dir/$vol_name.disk
rm -f $vol_file
sql_exec "delete from volume where name='$vol_name'"
echo "$vol_name|deleted"
