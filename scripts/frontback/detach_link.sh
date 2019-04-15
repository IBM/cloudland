#!/bin/bash

cd `dirname $0`
source ../cloudrc

[ $# -lt 2 ] && echo "$0 <router> <vlan>" && exit -1

router=$1
vlan=$2

sql_exec "update netlink set router='' where vlan='$vlan'"
vlans=`sql_exec "select vlans from router where name='$router'"`
vlans=`echo "$vlans" | sed "s/ $vlan//"`
sql_exec "update router set vlans='$vlans' where name='$router'"
