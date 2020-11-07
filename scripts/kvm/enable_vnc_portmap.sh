#!/bin/bash

cd `dirname $0`
source ../cloudrc

[ $# -lt 6 ] && die "$0 <vm_ID> <router_ID> <vnc_IP> <vnc_port> <access_IP> <access_port>"

ID=$1
vm_ID=inst-$1
rID=$2
router=router-$2
vnc_ip=$3
vnc_port=$4
access_ip=$5
access_port=$6

ip netns exec $router iptables -t nat -I PREROUTING -d $access_ip -p tcp --dport $access_port -j DNAT --to-destination $vnc_ip:$vnc_port
echo "ip netns exec $router iptables -t nat -D PREROUTING -d $access_ip -p tcp --dport $access_port -j DNAT --to-destination $vnc_ip:$vnc_port" | at now + 30 minutes
echo "|:-COMMAND-:| $(basename $0) '$ID' '${access_ip}' '${access_port}'"
