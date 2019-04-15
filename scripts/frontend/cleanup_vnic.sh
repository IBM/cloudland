#!/bin/bash

for i in $(ip -o link show | grep tap | cut -d: -f2 | cut -d@ -f1); do
    ip link del $i
done

for i in $(ip netns list | grep -v test | cut -d' ' -f1); do
    ip netns del $i
done

for i in $(ovs-vsctl show | grep Bridge | awk '{print $2}' | xargs); do
    ovs-vsctl del-br $i
done

for i in $(ip addr show bond0 | grep 172.250 | awk '{print $2}'); do
    ip addr del $i dev bond0
done
