#!/bin/bash

cd `dirname $0`
source ../cloudrc

[ $# -lt 4 ] && die "$0 <vm_ID> <router_ID> <vnc_IP> <vnc_port>"

ID=$1
vm_ID=inst-$1
rID=$2
router=router-$2
vnc_ip=$3
vnc_port=$4
public_ip=$(ip netns exec $router ifconfig te-$rID | grep 'inet ' | awk '{print $2}')
max_port=$(ip netns exec $router iptables -t nat -S | grep 'dport.* DNAT' | cut -d' ' -f10 | sort -u | tail -1)
map_port=18000
if [ -n "$max_port" ]; then
    let map_port=$max_port+$RANDOM%5
fi
ip netns exec $router iptables -t nat -I PREROUTING -d $vnc_ip -p tcp --dport $map_port -j DNAT --to-destination $vnc_ip:$vnc_port
echo "ip netns exec $router iptables -t nat -D PREROUTING -d $vnc_ip -p tcp --dport $map_port -j DNAT --to-destination $vnc_ip:$vnc_port" | at now + 30 minutes

echo "|:-COMMAND-:| $(basename $0) '$ID' '${public_ip}:${map_port}'"
