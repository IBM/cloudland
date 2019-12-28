#!/bin/bash

for inst in $(virsh list | grep inst- | awk '{print $2}'); do
    bridge=$(virsh dumpxml $inst | grep bridge= | cut -d"'" -f2 | xargs)
    mac_addr=$(virsh dumpxml $inst | grep 'mac address' | cut -d"'" -f2 | xargs)
    echo $inst $bridge $mac_addr $(hostname -s)
done
