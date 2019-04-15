#!/bin/bash

cd `dirname $0`
source ../cloudrc

[ $# -lt 2 ] && echo "$0 <router> <vlan>" && exit -1

router=$1
vlan=$2

sql_exec "update netlink set router='$router' where vlan='$vlan'"
sql_exec "update router set vlans=vlans||' $vlan' where name='$router'"
