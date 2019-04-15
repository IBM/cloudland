#!/bin/bash 

cd `dirname $0`
source ../cloudrc

[ $# -lt 2 ] && echo "$0 <user> <vlan> <network> [netmask] [gateway] [start_ip] [end_ip] [dhcp]" && exit -1

owner=$1
vlan=$2
network=$3
netmask=$4
gateway=$5
start_ip=$6
end_ip=$7
dhcp=$8

#[ "$vlan" -ge 4095 -o "$owner" == "admin" ] || die "Vlan number must be >= 4095"
[ -z "$dhcp" ] && dhcp=true

ipcalc -c "$network" >/dev/null 2>&1
[ $? -eq 0 ] || die "Invalid network!"
ipcalc -c "$netmask" >/dev/null 2>&1
[ $? -eq 0 ] || die "Invalid netmask!"
if [ -z "$gateway" ]; then
    let ip=`inet_aton $network`+1
    gateway=`inet_ntoa $ip`
fi
if [ "$gateway" != "nogateway" ]; then
    ipcalc -c "$gateway" >/dev/null 2>&1
    [ $? -eq 0 ] || die "Invalid gateway!"
    net=`ipcalc -n $gateway $netmask | cut -d'=' -f2`
    [ "$network" == "$net" ] || die "Invalid gateway!"
fi
[ "$gateway" == "start_ip" ] && die "Start IP can not Gateway IP"

if [ -n "$start_ip" ]; then
    net=`ipcalc -n $start_ip $netmask | cut -d'=' -f2`
    [ "$network" == "$net" ] || die "Invalid start_ip!"
else
    let ip=`inet_aton $network`+2
    start_ip=`inet_ntoa $ip`
fi
if [ -n "$end_ip" ]; then
    net=`ipcalc -n $end_ip $netmask | cut -d'=' -f2`
    [ "$network" == "$net" ] || die "Invalid end_ip!"
else
    let ip=`inet_aton $(ipcalc -b $network $netmask | cut -d'=' -f2)`-1
    end_ip=`inet_ntoa $ip`
fi

num=`sql_exec "select count(*) from netlink where owner='$owner' and vlan='$vlan'"`
[ $num -eq 1 ] || die "Not vlan $vlan owner"

sql_exec "insert into network (vlan, network, netmask, gateway, start_address, end_address) values ($vlan, '$network', '$netmask', '$gateway', '$start_ip', '$end_ip')" 
[ "$gateway" != "nogateway" ] && sql_exec "insert into address (IP, allocated, vlan, network) values ('$gateway', 'true', '$vlan', '$network')"
if [ -n "$network" ]; then
    sql_exec "insert into address (IP, allocated, vlan, network) values ('$start_ip', 'true', '$vlan', '$network')"

    ip=`inet_aton $start_ip`
    end=`inet_aton $end_ip`
    while [ $ip -lt $end ]; do
        let ip=$ip+1
        sql_exec "insert into address (IP, allocated, vlan, network) values ('`inet_ntoa $ip`', 'false', '$vlan', '$network')"
    done
fi

echo "$network|created"
