#!/bin/bash

for i in $(virsh list | grep running | awk '{print $2}'); do
    sudo /opt/cloudland/scripts/backend/clear_vm.sh $i
done

for i in $(ovs-vsctl show | grep Bridge | awk '{print $2}' | xargs); do
    ovs-vsctl del-br $i
done

for i in $(ip addr show bond0 | grep 172.250 | awk '{print $2}'); do
    ip addr del $i dev bond0
done
