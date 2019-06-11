#!/bin/bash

cd $(dirname $0)
source ../cloudrc

[ $# -lt 2 ] && die "$0 <number> <device>"

number=$1
device=$2

bridge=br$number
vxlan=v-$number

cat /proc/net/dev | grep -q "\<$bridge\>:"
[ $? -eq 0 ] && exit 1

addresses=$(ip addr show $device | grep 'inet ' | awk '{print $2}')
routes=$(ip route | grep $device | grep via | cut -d' ' -f1-3)
echo $addresses
echo "$routes"
nmcli connection delete bridge
nmcli connection add con-name $bridge type bridge ifname $bridge
for addr in $addresses; do
    nmcli connection modify $bridge +ipv4.addresses $addr ipv4.method static
done

while read line; do
    echo $line
    gateway=$(echo $line | cut -d' ' -f3)
    destination=$(echo $line | cut -d' ' -f1)
    echo $gateway
    echo $destination
    if [ "$destination" = "default" ]; then
        nmcli connection modify $bridge ipv4.gateway $gateway
    else
        nmcli connection modify $bridge ipv4.gateway $destination $gateway
    fi
done <<< "$routes"

nmcli connection down $device
nmcli connection modify $device ipv4.method disabled master $bridge
nmcli connection up $device
nmcli connection up $bridge
