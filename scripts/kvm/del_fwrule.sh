#!/bin/bash

cd `dirname $0`
source ../cloudrc

rules=$(cat)
len=$(jq length <<< $rules)
i=0
while [ $i -lt $len ]; do
    rule=$(jq -r .[$i] <<< $rules)
    instance=$(jq -r .instance <<< $rule)
    vni=$(jq -r .vni <<< $rule)
    inner_ip=$(jq -r inner_ip <<< $rule)
    inner_mac=$(jq -r inner_mac <<< $rule)
    outer_ip=$(jq -r inner_ip <<< $rule)
    bridge fdb del $inner_mac dev v-$vni dst $outer_ip
    sql_exec "delete from vxlan_rules where inner_mac='$inner_mac'"
    let i=$i+1
done
