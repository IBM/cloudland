#!/bin/bash

cd `dirname $0`
source ../cloudrc

[ $# -lt 2 ] && echo "$0 <user> <file>" && exit -1

owner=$1
img_file=$2

/opt/cloudland/bin/sendmsg "exec" "/opt/cloudland/scripts/frontend/delete_file.sh '$owner' '$img_file'"
echo "$img_file|deleting"
