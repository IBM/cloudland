#!/bin/bash -xv

cd $(dirname $0)
source ../cloudrc

[ $# -lt 2 ] && die "$0 <number> <device>"

number=$1
device=$2

bridge=br$number
vxlan=v-$number

../backend/clear_link.sh $number
cat /proc/net/dev | grep -q "\<$bridge\>:"
[ $? -eq 0 ] && exit 1

nmcli device set $device managed yes
conn=$(nmcli device show $device | grep GENERAL.CONNECTION | cut -d: -f2 | xargs)
addresses=$(nmcli connection show "$conn" | grep ipv4.addresses | cut -d: -f2 | xargs)
echo addr: $addresses
nmcli connection add con-name $bridge type bridge ifname $bridge
nmcli connection modify $bridge ipv4.addresses "$addresses" ipv4.method static
dns=$(nmcli connection show "$conn" | grep ipv4.dns: | cut -d: -f2 | xargs)
echo dns: $dns
nmcli connection modify $bridge ipv4.dns "$dns"
nmcli connection modify "$conn" -ipv4.dns "$dns"
gateway=$(nmcli connection show "$conn" | grep ipv4.gateway: | cut -d: -f2 | xargs)
echo gateway: $gateway
nmcli connection modify $bridge ipv4.gateway $gateway
nmcli connection modify "$conn" ipv4.gateway 0.0.0.0
nmcli connection modify "$conn" -ipv4.addresses "$addresses" ipv4.method disabled master $bridge
nmcli connection up "$conn"
nmcli connection up $bridge
