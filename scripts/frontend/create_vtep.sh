#!/bin/bash

cd `dirname $0`
source ../cloudrc

[ $# -lt 2 ] && die "$0 <instance> <vlan> [ip] [netmask] [mac]"

instance=$1
vlan=$2
ip=$3
netmask=$4
mac=$5

[ -z "$ip" ] && ip=$(sql_exec "select IP from address where vlan='$vlan' and allocated='false' limit 1")
[ -z "$ip" ] && die "No IP address is avalable"
sql_exec "update address set allocated='true' where IP='$ip'"
[ -z "$netmask" ] && netmask=$(sql_exec "select netmask from network where network=(select network from address where IP='$ip')")
[ -z "$mac" ] && mac="52:54:"$(openssl rand -hex 4 | sed 's/\(..\)/\1:/g; s/.$//')

sql_exec "insert into vtep (instance, vni, inner_ip, inner_mac, status) values ('$instance', '$vlan', '$ip', '$mac', 'inactive')"
/opt/cloudland/bin/grpcmsg "0" "inter= cpu=0 memory=0 disk=0" "/opt/cloudland/scripts/backend/`basename $0` '$instance' $vlan '$ip' '$netmask' '$mac'"
echo "$instance|creating"
