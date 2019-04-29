#!/bin/bash

cd `dirname $0`
source ../cloudrc

[ $# -lt 4 ] && die "$0 <gateway_id> <instance_ip> <instance_port> <remote_port>"

router=router-$1
inst_ip=${2%/*}
inst_port=$3
rport=$4

router_dir=/opt/cloudland/cache/router/$router
notify_sh=$router_dir/notify.sh
ssh_cmd="ip netns exec $router ssh -i /home/centos/.ssh/id_rsa -f -N -T -o StrictHostKeyChecking=no -R $rport:$inst_ip:$inst_port root@$portmap_remote_ip"
sed -i "\#$ssh_cmd#d" $notify_sh
ssh_pid=$(ip netns exec $router ps -ef | grep "ssh .* $rport:.*$portmap_remote_ip" | awk '{print $2}')
kill $ssh_pid
