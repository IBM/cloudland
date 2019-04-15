#!/bin/bash

cd `dirname $0`
source ../cloudrc

[ $# -lt 2 ] && echo "$0 <user> <router>" && exit -1

owner=$1
router=$2
num=`sql_exec "select count(*) from router where name='$router' and owner='$owner'"`
[ $num -lt 1 ] && die "Not the vlan $vlan owner!"
num=`sql_exec "select count(*) from netlink where router='$router'"`
[ $num -ge 1 ] && die "Router has attached VLAN(s)!"

sql_exec "update address set allocated='false' where instance='$router' and allocated='true'"
hyper_id=`sql_exec "select id from compute where hyper_name=(select host from router where name='$router')"`
sql_exec "delete from router where name='$router'"
/opt/cloudland/bin/sendmsg "inter $hyper_id" "/opt/cloudland/scripts/backend/`basename $0` $router"
echo "$router|deleted"
