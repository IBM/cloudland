#!/bin/bash

cd `dirname $0`
source ../cloudrc

[ $# -lt 1 ] && echo "$0 <vm_ID>" && exit -1

vm_ID=$1

net_json='
{
   "services" : [
      {
         "type" : "dns"
      }
  ],
   "networks": [],
   "links": []
}
'
net_json=$(echo $net_json | jq --arg dns $dns_server '.services[0].address |= .+$dns')
count=1

while read line; do
    [ -z "$line" ] && continue
    vlan=$(echo $line | cut -d' ' -f1)
    vm_ip=$(echo $line | cut -d' ' -f2)
    mac=$(echo $line | cut -d' ' -f3)
    decapper=$(echo $line | cut -d' ' -f4)
    primary=$(echo $line | cut -d' ' -f5)
    netmask=$(ipcalc -m $vm_ip | cut -d= -f2)
    addr=${vm_ip%/*}
	gateway=$(ipcalc --minaddr $vm_ip | cut -d= -f2)
	if [ "$primary" = 'true' ]; then
		net_json=$(echo $net_json | jq --arg netmask $netmask --arg gateway $gateway --arg addr $addr --arg mac $mac \
			'.networks[0].type="ipv4" | .networks[0].netmask=$netmask | .networks[0].link="eth0" | .networks[0].ip_address=$addr | .networks[0].id="network0" |
			 .networks[0].routes[0].network="0.0.0.0" | .networks[0].routes[0].netmask="0.0.0.0" | .networks[0].routes[0].gateway=$gateway | 
			 .links[0].ethernet_mac_address=$mac | .links[0].mtu=1450 | .links[0].id="eth0" | .links[0].type="phy"')
	else
		net_json=$(echo $net_json | jq --argjson id $count --arg netmask $netmask --arg addr $addr --arg mac $mac --arg netid network$count --arg linkid eth$count  \
			'.networks[$id].type="ipv4" | .networks[$id].netmask=$netmask | .networks[$id].link=$linkid | .networks[$id].ip_address=$addr | .networks[$id].id=$netid |
			 .links[$id].ethernet_mac_address=$mac | .links[$id].mtu=1450 | .links[$id].id=$linkid | .links[$id].type="phy"')
		let count=$count+1
	fi
    vif_name=nic$(echo $mac | cut -d: -f3- | tr -d ':')
    if [ "$use_lb" = "false" ]; then
        br_name=br$SCI_CLIENT_ID
        [ -z "$tunip" ] && get_tunip
        ovs-vsctl --may-exist add-br $br_name
        if_xml=$xml_dir/$vm_ID/$vif_name.xml
        cp $template_dir/interface.xml $if_xml
        sed -i "s#VM_MAC#$mac#g; s#VM_BRIDGE#$br_name#g; s/VM_VTEP/$vif_name/g;" $if_xml
        virsh attach-device $vm_ID $if_xml --config
        cmd="icp-tower --ovs-bridge=$br_name gate add --direct-routing --encap-identifier $vlan --local-ip=$tunip --interface $vif_name --vsi-mac-address $mac --vsi-ip-prefix ${vm_ip} --decapper-ip $decapper"
        sidecar span log $span "Internal: $cmd" "Result: $result"
        result=$(eval "$cmd")
    else
        br_name=br$vlan
        ./create_link.sh $vlan
        virsh attach-interface $vm_ID bridge $br_name --model virtio --mac $mac --target $vif_name --config
        ./create_sg_chain.sh $vif_name $addr $mac
    fi
    [ $? -eq 0 ] && echo "NIC $mac in vlan $vlan was attached successfully to $vm_ID."
done

vm_xml=$xml_dir/$vm_ID/$vm_ID.xml
virsh dumpxml --security-info $vm_ID 2>/dev/null | sed "s/autoport='yes'/autoport='no'/g" > $vm_xml.dump && mv -f $vm_xml.dump $vm_xml
latest_dir=/tmp/$vm_ID/openstack/latest
mkdir -p $latest_dir 2>/dev/null
echo "$net_json" > $latest_dir/network_data.json
echo "|:-COMMAND-:| launch_vm.sh '$vm_ID' 'running' '$SCI_CLIENT_ID'"
