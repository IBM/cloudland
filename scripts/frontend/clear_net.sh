#!/bin/bash

cd `dirname $0`
source ../cloudrc

[ $# -lt 2 ] && echo "$0 <user> <vlan> <network>" && exit -1

owner=$1
vlan=$2
network=$3

num=`sql_exec "select count(*) from netlink where vlan='$vlan' and owner='$owner'"`
[ $num -lt 1 ] && die "Not the vlan owner!"
num=`sql_exec "select count(*) from instance where vlan='$vlan' and status='running'"`
[ $num -ge 1 ] && die "The network is being used by instance(s)!"
sql_ret=`sql_exec "select gateway, start_address, netmask from network where network='$network'"`
gateway=`echo $sql_ret | cut -d'|' -f1`
start_addr=`echo $sql_ret | cut -d'|' -f2`
netmask=`echo $sql_ret | cut -d'|' -f3`
ip_addrs=`sql_exec "select IP from address where vlan='$vlan' and allocated='true'"`
for i in $ip_addrs; do
    [ "$i" == "$start_addr" -o "$i" == "$gateway" ] && continue;
    net=`ipcalc -n $i $netmask | cut -d'=' -f2`
    [ "$network" == "$net" ] && die "The network has address(es) in use!"
done

tag_id=`sql_exec "select id from network where network='$network' and vlan='$vlan'"`
sql_exec "delete from network where id='$tag_id'"
sql_exec "delete from address where network='$network' and vlan='$vlan'"

dh_host=`sql_exec "select dh_host from netlink where vlan='$vlan'"`
hyper_id=`sql_exec "select id from compute where hyper_name='$dh_host'"`
[ $hyper_id -ge 0 ] && /opt/cloudland/bin/sendmsg "inter $hyper_id" "/opt/cloudland/scripts/backend/`basename $0` $vlan $tag_id"
router=`sql_exec "select router from netlink where vlan='$vlan'"`
if [ -n "$router" ]; then
    hyper_id=`sql_exec "select id from compute where hyper_name=(select host from router where name='$router')"`
    /opt/cloudland/bin/sendmsg "inter $hyper_id" "/opt/cloudland/scripts/backend/clear_gateway.sh $router $vlan $gateway $netmask"
fi
echo "$network|deleted"
