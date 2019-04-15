#!/bin/bash

cd `dirname $0`
source ../cloudrc

[ $# -lt 3 ] && echo "$0 <user> <router> <vlan> <network>" && exit -1

owner=$1
router=$2
vlan=$3
network=$4

num=`sql_exec "select count(*) from router where name='$router' and owner='$owner'"`
[ $num -lt 1 ] && die "Not the router $router owner!"
num=`sql_exec "select count(*) from netlink where vlan='$vlan' and owner='$owner'"`
[ $num -lt 1 ] && die "Not the vlan $vlan owner!"
rt=`sql_exec "select router from netlink where vlan='$vlan' and owner='$owner'"`
[ "$rt" != "$router" ] && die "Vlan $vlan is attached to $rt, not $router"

sql_ret=`sql_exec "select gateway,netmask from network where vlan='$vlan' and network='$network'"`
[ -z "$sql_ret" ] && die "Vlan $vlan hasn't network $network"
gateway=`echo $sql_ret | cut -d'|' -f1`
netmask=`echo $sql_ret | cut -d'|' -f2`
hyper_id=`sql_exec "select id from compute where hyper_name=(select host from router where name='$router')"`
/opt/cloudland/bin/sendmsg "inter $hyper_id" "/opt/cloudland/scripts/backend/`basename $0` '$router' '$vlan' '$gateway' '$netmask'"
echo "$router|$vlan|routed"
