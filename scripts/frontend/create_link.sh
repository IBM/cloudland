#!/bin/bash

cd `dirname $0`
source ../cloudrc

[ $# -lt 1 ] && echo "$0 <user> [shared(true|false)] [description]" && exit -1

owner=$1
shared=$2
desc=$3

num=`sql_exec "select count(*) from netlink where owner='$owner'"`
quota=`sql_exec "select net_limit from quota where role=(select role from users where username='$owner')"`
[ $quota -ge 0 -a $num -ge $quota ] && die "Your quota is used up!"

vlan=`sql_exec "select MAX(vlan) from netlink"`
let vlan=$vlan+1

[ -z "$shared" ] && shared='false'
sql_exec "insert into netlink(vlan, owner, dh_host, shared, description) values ($vlan, '$owner', 'NO_DHCP', '$shared', '$desc')" 

echo "$vlan|created"
