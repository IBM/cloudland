ovs-ofctl add-flow $br_name "table=0,priority=99,in_port=2,dl_type=0x0806,nw_dst=172.16.11.82,actions=move:NXM_OF_ETH_SRC[]->NXM_OF_ETH_DST[],mod_dl_src:52:54:52:ec:ce:c1,load:0x2->NXM_OF_ARP_OP[],move:NXM_OF_ARP_TPA[]->NXM_OF_ARP_SPA[],move:NXM_NX_ARP_SHA[]->NXM_NX_ARP_THA[], move:NXM_OF_ARP_SPA[]->NXM_OF_ARP_TPA[], load:0x525452eccec1->NXM_NX_ARP_SHA[], in_port"
ovs-ofctl add-flow $br_name "
       table=0,
       priority=90,
       in_port=2,
       ip,
       nw_dst=172.16.11.82,
       actions=set_field:100100->tun_id,
       set_field:192.168.1.110->tun_dst,10"
ovs-ofctl add-flow $br_name "
       table=0,
       priority=90,
       in_port=10,
       ip,
       tun_id=100100,
       actions=output:2"
ovs-ofctl add-flow $br_name "table=0,priority=85,in_port=10,tun_id=100100,actions=normal"

ovs-ofctl add-flow $br_name "table=0,priority=80,arp,actions=set_field:100100->tun_id,set_field:192.168.1.125->tun_dst,10"
ovs-ofctl add-flow $br_name "table=0,priority=85,in_port=10,tun_id=100100,actions=learn(table=10,NXM_OF_ETH_DST[]=NXM_OF_ETH_SRC[],load:NXM_NX_TUN_ID[]->NXM_NX_TUN_ID[],load:NXM_NX_TUN_IPV4_SRC[]->NXM_NX_TUN_IPV4_DST[],output=NXM_OF_IN_PORT[]),normal"
ovs-ofctl add-flow $br_name "
       table=0,
       priority=100,
       in_port=2,ip,actions=resubmit(,10)"
