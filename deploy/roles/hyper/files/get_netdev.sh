#!/bin/bash

local_ip=$1
netdev=$(ip addr show | grep $local_ip | sed 's/.* //')
if [ "${netdev##br}" != "$netdev" ]; then
    netdev=$(ip -d -o link show | grep 'master br5000' | grep bridge_slave | head -1 | cut -d: -f2 | xargs)
fi
echo "$netdev"
