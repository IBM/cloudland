#!/bin/bash

cd `dirname $0`
source ../cloudrc

rules=$(cat)
len=$(jq length <<< $rules)
vtep_ip=$(ifconfig $vxlan_interface | grep 'inet ' | awk '{print $2}')
i=0
while [ $i -lt $len ]; do
    rule=$(jq -r .[$i] <<< $rules)
#    instance=$(jq -r .instance <<< $rule)
    router=$(jq -r .router <<< $rule)
    vni=$(jq -r .vni <<< $rule)
    cat /proc/net/dev | grep -q "\<br$vni\>:"
    if [ $? -ne 0 ]; then
        ./create_link.sh $vni
        gateway=$(jq -r .gateway <<< $rule)
        ./set_subnet_gw.sh $router $vni $gateway
    fi
    inner_ip=$(jq -r .inner_ip <<< $rule)
    in_ip=${inner_ip%%/*}
    inner_mac=$(jq -r .inner_mac <<< $rule)
    vm_name=$(jq -r .hostname <<< $rule)
    [ -n "$vm_name" ] && ./set_host.sh "$router" "$vni" "$inner_mac" "$vm_name" "$in_ip"

    outer_ip=$(jq -r .outer_ip <<< $rule)
    if [ "$outer_ip" != "$vtep_ip" ]; then
	bridge fdb | grep "\<$inner_mac\>"
        [ $? -eq 0 ] && bridge fdb del $inner_mac dev v-$vni
        bridge fdb add $inner_mac dev v-$vni dst $outer_ip self permanent
	ip neighbor | grep "\<$in_ip\> dev v-$vni\>"
        [ $? -eq 0 ] && ip neighbor del $in_ip dev v-$vni
        ip neighbor add $in_ip lladdr $inner_mac dev v-$vni nud permanent
    fi
    let i=$i+1
done
