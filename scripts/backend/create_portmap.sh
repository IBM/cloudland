#!/bin/bash

cd `dirname $0`
source ../cloudrc

[ $# -lt 4 ] && die "$0 <gateway_id> <instance_ip> <instance_port> <remote_port>"

router=router-$1
inst_ip=$2
inst_port=$3
rport=$4

ip netns exec $router ssh -i /home/centos/.ssh/id_rsa -o StrictHostKeyChecking=no -R $rport:$inst_ip:$inst_port root@$portmap_remote_ip
