#!/bin/bash

cd `dirname $0`
source ../cloudrc

rules=$(cat)
len=$(jq length <<< $rules)
vtep_ip=$(ifconfig $vxlan_interface | grep 'inet ' | awk '{print $2}')
i=0
while [ $i -lt $len ]; do
#    instance=$(jq -r .instance <<< $rule)
    read -d'\n' -r vni router gateway outer_ip inner_ip inner_mac < <(jq -r ".[$i].vni, .[$i].router, .[$i].gateway, .[$i].outer_ip, .[$i].inner_ip, .[$i].inner_mac" <<<$rules)
    cat /proc/net/dev | grep -q "\<br$vni\>:"
    if [ $? -ne 0 ]; then
        ./create_link.sh $vni
        ./set_subnet_gw.sh $router $vni $gateway
    fi
    if [ "$outer_ip" != "$vtep_ip" ]; then
	bridge fdb | grep "\<$inner_mac\>"
        [ $? -eq 0 ] && bridge fdb del $inner_mac dev v-$vni
        bridge fdb add $inner_mac dev v-$vni dst $outer_ip self permanent
	in_ip=${inner_ip%%/*}
	ip neighbor | grep "\<$in_ip\> dev v-$vni\>"
        [ $? -eq 0 ] && ip neighbor del $in_ip dev v-$vni
        ip neighbor add $in_ip lladdr $inner_mac dev v-$vni nud permanent
    fi
    let i=$i+1
done
