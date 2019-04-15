#!/bin/bash

cd `dirname $0`
source ../cloudrc

[ $# -lt 1 ] && echo "$0 <user> [outgoing(true|false)] [shared(true|false)] [description]" && exit -1

owner=$1
outgoing=$2
shared=$3
desc=$4

num=`sql_exec "select count(*) from router where owner='$owner'"`
quota=`sql_exec "select router_limit from quota where role=(select role from users where username='$owner')"`
[ $quota -ge 0 -a $num -ge $quota ] && die "Your quota is used up!"

[ -z "$shared" ] && shared='false'
[ -z "$outgoing" ] && outgoing='false'
if [ "$outgoing" == 'true' ]; then
    sql_ret=`sql_exec "select IP, network from address where vlan='$external_vlan' and allocated='false' limit 1"`
    out_ip=`echo $sql_ret | cut -d'|' -f1`
    network=`echo $sql_ret | cut -d'|' -f2`
    netmask=`sql_exec "select netmask from network where network='$network' and vlan='$external_vlan'"`
    out_mac="52:50:"`openssl rand -hex 4 | sed 's/\(..\)/\1:/g; s/.$//'`
    sql_exec "update address set allocated='true', mac='$out_mac', instance='$router' where IP='$out_ip'"
fi
router="router-"`openssl rand -hex 4`

sql_exec "insert into router (name, description, owner, vlans, out_addr, shared) values ('$router', '$desc', '$owner', '', '$out_ip', '$shared')" 
/opt/cloudland/bin/sendmsg "inter 0" "/opt/cloudland/scripts/backend/`basename $0` $router $out_ip $netmask"

echo "$router|created"
