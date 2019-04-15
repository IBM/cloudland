#!/bin/bash

cd `dirname $0`
source ../cloudrc

[ $# -lt 1 ] && echo "$0 <comp_no>" && exit -1

exit 0

comp=$1
stat=`sql_exec "select status from compute where id='$comp'"`
[ "$stat" == "dead" ] && exit 0

sql_exec "update compute set status='dead' where id='$comp'"
vlan_list=`sql_exec "select vlan from netlink where dh_host=(select hyper_name from compute where id='$comp')"`
for vlan in $vlan_list; do
    sql_ret=`sql_exec "select network,netmask,start_address from network where vlan='$vlan'"`
    network=`echo $sql_ret | cut -d'|' -f1`
    netmask=`echo $sql_ret | cut -d'|' -f2`
    start_ip=`echo $sql_ret | cut -d'|' -f3`
    /opt/cloudland/bin/sendmsg "inter" "/opt/cloudland/scripts/backend/create_net.sh '$vlan' '$network' '$netmask' '' '$start_ip' '' '' 'yes'"
    echo "$vlan|rescued"
done

router_list=`sql_exec "select name,out_ip from router where host=(select hyper_name from compute where id='$comp')"`
for rinfo in $router_list; do
    router=`echo $rinfo | cut -d'|' -f1`
    out_ip=`echo $rinfo | cut -d'|' -f2`
    /opt/cloudland/bin/sendmsg "inter" "/opt/cloudland/scripts/backend/create_router.sh $router $out_ip $netmask"
    sleep 5
    hyper_id=`sql_exec "select id from compute where hyper_name=(select host from router where name='$router')"`
    vlans=`sql_exec "select vlan from netlink where router='$router'"`
    for vlan in $vlans; do
        gateways=`sql_exec "select gateway,netmask from network where vlan='$vlan'"`
        /opt/cloudland/bin/sendmsg "inter $hyper_id" "/opt/cloudland/scripts/backend/`basename $0` '$router' '$vlan' '$gateways'"
    done
    echo "$router|rescued"
done
