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
    vni=$(jq -r .vni <<< $rule)
    cat /proc/net/dev | grep -q "\<br$vni\>:"
    if [ $? -ne 0 ]; then
        ./create_link.sh $vni
        router=$(jq -r .router <<< $rule)
        gateway=$(jq -r .gateway <<< $rule)
        ./set_subnet_gw.sh $router $vni $gateway
    fi
    outer_ip=$(jq -r .outer_ip <<< $rule)
    if [ "$outer_ip" != "$vtep_ip" ]; then
        inner_ip=$(jq -r .inner_ip <<< $rule)
        inner_mac=$(jq -r .inner_mac <<< $rule)
	bridge fdb del $inner_mac dev v-$vni
        bridge fdb add $inner_mac dev v-$vni dst $outer_ip self permanent
	ip neighbor del ${inner_ip%%/*} dev v-$vni
        ip neighbor add ${inner_ip%%/*} lladdr $inner_mac dev v-$vni nud permanent
    fi
#    sql_exec "insert into vxlan_rules (instance, vni, inner_ip, inner_mac, outer_ip) values ('$instance', '$vni', '$inner_ip', '$inner_mac', '$outer_ip')"
    let i=$i+1
done
