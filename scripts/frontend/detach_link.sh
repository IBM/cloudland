#!/bin/bash

cd `dirname $0`
source ../cloudrc

[ $# -lt 3 ] && echo "$0 <user> <router> <vlan>" && exit -1

owner=$1
router=$2
vlan=$3

num=`sql_exec "select count(*) from router where name='$router' and owner='$owner'"`
[ $num -lt 1 ] && die "Not the router $router owner!"
num=`sql_exec "select count(*) from netlink where vlan='$vlan' and owner='$owner'"`
[ $num -lt 1 ] && die "Not the vlan $vlan owner!"
rt=`sql_exec "select router from netlink where vlan='$vlan'"`
[ "$router" != "$rt" ] && die "Vlan $vlan not attached to router $router"

gateways=`sql_exec "select gateway,netmask from network where vlan='$vlan'"`
hyper_id=`sql_exec "select id from compute where hyper_name=(select host from router where name='$router')"`
/opt/cloudland/bin/sendmsg "inter $hyper_id" "/opt/cloudland/scripts/backend/`basename $0` '$router' '$vlan' '$gateways'"
echo "$router|$vlan|detached"
