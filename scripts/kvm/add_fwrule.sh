#!/bin/bash -xv

cd `dirname $0`
source ../cloudrc

rules=$(cat)
len=$(jq length <<< $rules)
i=0
while [ $i -lt $len ]; do
    rule=$(jq -r .[$i] <<< $rules)
    instance=$(jq -r .instance <<< $rule)
    vni=$(jq -r .vni <<< $rule)
    inner_ip=$(jq -r .inner_ip <<< $rule)
    inner_mac=$(jq -r .inner_mac <<< $rule)
    outer_ip=$(jq -r .outer_ip <<< $rule)
    bridge fdb add $inner_mac dev v-$vni dst $outer_ip self permanent
    ip neighbor add ${inner_ip%%/*} lladdr $inner_mac dev v-$vni nud permanent
    sql_exec "insert into vxlan_rules (instance, vni, inner_ip, inner_mac, outer_ip) values ('$instance', '$vni', '$inner_ip', '$inner_mac', '$outer_ip')"
    let i=$i+1
done
