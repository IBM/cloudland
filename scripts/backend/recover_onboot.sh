#!/bin/bash

cd $(dirname $0)
source ../cloudrc

for conf in $cache_dir/router/*; do
    router=$(basename $conf)
    ./load_keepalived_conf.py -q $conf/keepalived.conf
    udevadm settle
    ip netns exec $router keepalived -D -f $conf/keepalived.conf -p $conf/keepalived.pid -r $conf/vrrp.pid -c $conf/checkers.pid
done

for inst in $(virsh list --all | grep 'shut off' | awk '{print $2}'); do
    virsh start $inst
done

exit 0
