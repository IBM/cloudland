#!/bin/bash

cd `dirname $0`
source ../cloudrc

[ $# -lt 1 ] && die "$0 <instance>"

instance=$1

ip netns exec $instance ip link set lo down
vm_nic=$(ip netns exec $instance ip -o link show | cut -d: -f2 | cut -d@ -f1 | grep vm- | xargs)
if [ -n "$vm_nic" ]; then
    vs_nic=${vm_nic/vm-/tap}
    br_name=br$SCI_CLIENT_ID
    icp-tower --ovs-bridge=$br_name gate remove --interface $vs_nic
    ip link del $vs_nic
    ip netns del $instance
    rm -f $cache_dir/meta/$vm_ID.sh
#    sidecar span log $span "Internal: namespace $instance and vnic $vs_nic is deleted" "result: $result" "Callback: clear_vnic.sh '$vif_dev'"
    echo "|:-COMMAND-:| clear_vnic.sh '$vs_nic'"
fi
#sidecar span log $span "Callback: `basename $0` '$instance'"
echo "|:-COMMAND-:| `basename $0` '$instance'"
