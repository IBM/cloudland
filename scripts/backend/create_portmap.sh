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
ssh_cmd="ip netns exec $router ssh -i $cland_private_key -f -n -N -T -o StrictHostKeyChecking=no -R $rport:$inst_ip:$inst_port cland@$portmap_remote_ip"
sed -i "\#$ssh_cmd#d" $notify_sh
sed -i "/\"MASTER\")/a $ssh_cmd" $notify_sh
eval $ssh_cmd &>/dev/null
[ $? -eq 0 ] && echo "|:-COMMAND-:| `basename $0` '$portmap_remote_ip' '$rport'"
