#!/bin/bash

device=$(ip route | grep default | awk '{print $5}')
[ "${device##br}" != "$device" ] && exit 0

ether_cfg=/etc/sysconfig/network-scripts/ifcfg-$device
bridge_cfg=/etc/sysconfig/network-scripts/ifcfg-br5000
grep -q 'BRIDGE=' $ether_cfg
[ $? -eq 0 ] && exit 0

gateway=$(grep "^GATEWAY" $ether_cfg)
ipaddr=$(grep "^IPADDR" $ether_cfg)
netmask=$(grep "^NETMASK" $ether_cfg)
prefix=$(grep "^PREFIX" $ether_cfg)

[ -n "$gateway" ] && sed -i "/$gateway/d" $ether_cfg
[ -n "$ipaddr" ] && sed -i "/$ipaddr/d" $ether_cfg
[ -n "$netmask" ] && sed -i "/$netmask/d" $ether_cfg
[ -n "$prefix" ] && sed -i "/$prefix/d" $ether_cfg
echo "BRIDGE=br5000" >> $ether_cfg

cat >$bridge_cfg <<EOF
STP=yes
TYPE=Bridge
PROXY_METHOD=none
BROWSER_ONLY=no
BOOTPROTO=none
DEFROUTE=yes
IPV4_FAILURE_FATAL=no
IPV6INIT=yes
IPV6_AUTOCONF=yes
IPV6_DEFROUTE=yes
IPV6_FAILURE_FATAL=no
IPV6_ADDR_GEN_MODE=stable-privacy
NAME=br5000
DEVICE=br5000
ONBOOT=yes
EOF
[ -n "$gateway" ] && echo $gateway >> $bridge_cfg
[ -n "$ipaddr" ] && echo $ipaddr >> $bridge_cfg
[ -n "$netmask" ] && echo $netmask >> $bridge_cfg
[ -n "$prefix" ] && echo $prefix >> $bridge_cfg
