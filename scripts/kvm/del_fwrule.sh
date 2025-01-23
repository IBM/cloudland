#!/bin/bash

cd `dirname $0`
source ../cloudrc

rules=$(cat)
len=$(jq length <<< $rules)
vtep_ip=$(ifconfig $vxlan_interface | grep 'inet ' | awk '{print $2}')
i=0
while [ $i -lt $len ]; do
    rule=$(jq -r .[$i] <<< $rules)
    router=$(jq -r .router <<< $rule)
#    instance=$(jq -r .instance <<< $rule)
    vni=$(jq -r .vni <<< $rule)
    inner_ip=$(jq -r .inner_ip <<< $rule)
    inner_mac=$(jq -r .inner_mac <<< $rule)
    outer_ip=$(jq -r .outer_ip <<< $rule)
    bridge fdb del $inner_mac dev v-$vni
    ip neighbor del ${inner_ip%%/*} dev v-$vni
    if [ "$outer_ip" != "$vtep_ip" ]; then
        ./del_host.sh "$router" "$vni" "$inner_mac" "$inner_mac"
    fi
    let i=$i+1
done
