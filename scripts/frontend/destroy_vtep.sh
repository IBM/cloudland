#!/bin/bash

cd `dirname $0`
source ../cloudrc

[ $# -lt 1 ] && die "$0 <instance>"

instance=$1
hyper_id=$(sql_exec "select hyper_id from vtep where instance='$instance'")
/opt/cloudland/bin/grpcmsg "0" "inter=$hyper_id" "/opt/cloudland/scripts/backend/`basename $0` '$instance'"
echo "$instance|destroying"
