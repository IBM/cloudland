#!/bin/bash

cd $(dirname $0)
source ../cloudrc

for conf in $cache_dir/router/router-*; do
    router=$(basename $conf)
    ip netns add $router
    ip netns exec $router bash -c "echo 1 >/proc/sys/net/ipv4/ip_forward"
    ip netns exec $router bash -c "/sbin/ipset restore < $conf/ipset.save"
    ip netns exec $router bash -c "/sbin/iptables-restore < $conf/iptables.save"
    ./load_keepalived_conf.py -q $conf/keepalived.conf
    udevadm settle
    ip netns exec $router keepalived -D -f $conf/keepalived.conf -p $conf/keepalived.pid -r $conf/vrrp.pid -c $conf/checkers.pid
done

for conf in $cache_dir/dnsmasq/vlan*; do
    for cmd in $conf/tags/*/cmd; do
        sh $cmd
    done
done

for inst in $(virsh list --all | grep 'shut off' | awk '{print $2}'); do
    virsh start $inst
done

#add mac to fdb after recovering
inface1=$vlan_interface
instances=`virsh list --all | grep inst | awk '{print $2}'`
for instance in $instances
do
   macs=`virsh domiflist $instance | grep 52:| awk '{print $5}'`
   echo "$instance MAC addresses are: $macs" >> /tmp/cloudland/recover_onboot.log
   for mac in $macs
   do
        echo "Registing $mac" >> /tmp/cloudland/recover_onboot.log
        /usr/sbin/bridge fdb add $mac dev $inface1 2>/dev/null
   done
   echo "br-ex1 fbd sync done for $instance"
done


exit 0
