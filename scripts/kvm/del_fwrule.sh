#!/bin/bash

cd `dirname $0`
source ../cloudrc

rules=$(cat)
len=$(jq length <<< $rules)
vtep_ip=$(ifconfig $vxlan_interface | grep 'inet ' | awk '{print $2}')
i=0
while [ $i -lt $len ]; do
    read -d'\n' -r vni router outer_ip inner_ip inner_mac < <(jq -r ".[$i].vni, .[$i].router, .[$i].outer_ip, .[$i].inner_ip, .[$i].inner_mac" <<<$rules)
    bridge fdb del $inner_mac dev v-$vni
    ip neighbor del ${inner_ip%%/*} dev v-$vni
    if [ "$outer_ip" = "$vtep_ip" ]; then
        ./del_host.sh "$router" "$vni" "$inner_mac" "$inner_mac"
    fi
    let i=$i+1
done
