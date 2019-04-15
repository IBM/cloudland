#!/bin/bash

cd `dirname $0`
source ../cloudrc

[ $# -lt 2 ] && echo "$0 <user> <vlan>" && exit -1

owner=$1
vlan=$2
num=`sql_exec "select count(*) from netlink where vlan='$vlan' and owner='$owner'"`
[ $num -lt 1 ] && die "Not the vlan $vlan owner!"
num=`sql_exec "select count(*) from network where vlan='$vlan'"`
[ $num -ge 1 ] && die "Vlan has network!"
rt=`sql_exec "select router from netlink where vlan='$vlan' and owner='$owner'"`
[ -n "$rt" ] && die "Vlan $vlan is attached to router $rt"

sql_exec "delete from netlink where vlan='$vlan'"

hyper_id=`sql_exec "select id from compute where hyper_name=(select dh_host from netlink where vlan='$vlan')"`
/opt/cloudland/bin/sendmsg "inter $hyper_id" "/opt/cloudland/scripts/backend/`basename $0` $vlan"
echo "$vlan|deleted"
