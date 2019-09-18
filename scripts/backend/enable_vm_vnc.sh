#!/bin/bash

cd `dirname $0`
source ../cloudrc

vm_ID=inst-$1
vnc_ip=$3
vnc_port=18000
vnc_pass=password
ip netns exec router-$2 iptables -t nat -I PREROUTING -d $vnc_ip -p tcp --dport $vnc_port -j DNAT --to-destination $portmap_remote_ip:5900
echo "|:-COMMAND-:| $(basename $0) '6' '$vnc_ip' '$vnc_port' '$vnc_pass'"
