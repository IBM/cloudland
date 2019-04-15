#!/bin/bash

cd `dirname $0`
source ../cloudrc

[ $# -lt 1 ] && echo "$0 <user>" && exit -1

owner=$1
sql_exec "select name, size, description, inst_id, device, bootable, status from volume where owner='$owner'"
