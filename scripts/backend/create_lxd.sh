#!/bin/bash

cd `dirname $0`
source ../cloudrc

[ $# -lt 6 ] && die "$0 <vm_ID> <image> <name> <flavor> <vlan> <cidr> <gateway> [userdata] [pubkey]"

ID=$1
image=$2
name=$3
flavor=$4
vlan=$5
cidr=$6
gateway=$7
userdata=$8
pubkey=$9

lxname=${name}-${ID%%-*}
bridge=br$vlan
./create_link.sh $vlan

lxc init $image $lxname -p template
lxc config device add $lxname eth0 nic nictype=bridged parent=$bridge
netfile=$container_dir/$lxname/templates/cloud-init-network.tpl
cat > $netfile <<EOF
version: 1
config:
    - type: physical
      name: eth0
      subnets:
          - type: static
            ipv4: true
            address: ${cidr%/*}
            netmask: $(ipcalc -m $cidr | cut -d= -f2)
            gateway: $gateway
            control: auto
    - type: nameserver
      address: $dns_server
EOF
lxc start $lxname
