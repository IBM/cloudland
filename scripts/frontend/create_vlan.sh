#!/bin/bash

cd `dirname $0`
source ../cloudrc

[ $# -lt 3 ] && echo "$0 <user> <vlan> <vpc_name> [shared(true|false)] [description]" && exit -1

owner=$1
vlan=$2
vpc_name=$3
shared=$4
desc=$5

#[ "$vlan" -ge 4095 -o "$owner" == "admin" ] || die "Vlan number must be >= 4095"
[ "$vlan" -ge 4095 -o "$owner" == "$admin_user" ] || die "Vlan number must be >= 4095"

num=`sql_exec "select count(*) from netlink where vlan='$vlan'"`
[ $num -eq 0 ] || die "Vlan alreay exists!"
num=`sql_exec "select count(*) from netlink where owner='$owner'"`

[ -z "$shared" ] && shared='false'
sql_exec "insert into netlink(vlan, owner, vpc_name, shared, description) values ($vlan, '$owner', '$vpc_name', '$shared', '$desc')" 

echo "$vlan|created"
