#!/bin/bash

cd `dirname $0`
source ../cloudrc

[ $# -lt 1 ] && echo "$0 <user>" && exit -1

owner=$1
vlans=`sql_exec "select vlan from netlink where owner='$owner' or shared='true' COLLATE NOCASE"`
for vlan in $vlans; do
    sql_exec "select network, netmask, gateway, start_address, end_address, vlan from network where vlan='$vlan'"
done
