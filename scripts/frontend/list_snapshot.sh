#!/bin/bash

cd `dirname $0`
source ../cloudrc

[ $# -lt 1 ] && echo "$0 <user>" && exit -1

owner=$1
sql_exec "select inst_id, status, description, owner from snapshot where status!='deleted'"

