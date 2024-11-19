#!/bin/bash

cd `dirname $0`
source ../cloudrc

rules=$(cat)
len=$(jq length <<< $rules)
i=0
while [ $i -lt $len ]; do
    rule=$(jq -r .[$i] <<< $rules)
    router=router-$(jq -r .router <<< $rule)
    ip netns | grep "\<$router\>"
    [ $? -eq 0 ] && continue
#    instance=$(jq -r .instance <<< $rule)
    vni=$(jq -r .vni <<< $rule)
    inner_ip=$(jq -r .inner_ip <<< $rule)
    inner_mac=$(jq -r .inner_mac <<< $rule)
#    outer_ip=$(jq -r .outer_ip <<< $rule)
    bridge fdb del $inner_mac dev v-$vni
    ip neighbor del ${inner_ip%%/*} dev v-$vni
#    sql_exec "delete from vxlan_rules where inner_mac='$inner_mac'"
    let i=$i+1
done
