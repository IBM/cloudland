#!/bin/bash

cd $(dirname $0)
source ../cloudrc

[ $# -lt 1 ] && echo "$0 <router>" && exit -1

ID=$1
router=router-$ID

[ -z "$router" ] && exit 1

ns_links=$(ip netns exec $router ip link show | grep ns- | awk '{print $2}' | cut -d'@' -f1)
for link in $ns_links; do
    vlan=${link/ns-/}
    slaves=$(ls -A /sys/devices/virtual/net/br$vlan/brif | grep -v "v-\|ln-")
    if [ -n "$slaves" ]; then
        echo "There are active vms"
        exit 0
    fi
done

for link in $ns_links; do
    vlan=${link/ns-/}
    ./clear_link.sh $vlan
done

ip netns exec $router ip link set lo down
suffix=$ID
ip link del ext-$suffix
ip link del int-$suffix
ip netns del $router

nat_ip=169.$(($SCI_CLIENT_ID % 234)).$(($suffix % 234)).3
route_ip=$(ifconfig $vxlan_interface | grep 'inet ' | awk '{print $2}')
iptables -t nat -D POSTROUTING -s ${nat_ip}/32 -j SNAT --to-source $route_ip
