#!/bin/bash

cd $(dirname $0)
source ../cloudrc

[ $# -lt 5 ] && echo "$0 <router> <ext_defaut_gw> <int_defaut_gw> <ext_gw_cidr> <int_gw_cidr> <vrrp_vni> <vrrp_ip> <role>" && exit -1

ID=$1
router=router-$ID
ext_gw=${2%/*}
int_gw=${3%/*}
ext_ip=$4
int_ip=$5
vrrp_vni=$6
vrrp_ip=$7
role=$8

[ -z "$router" -o -z "$ext_ip" -o -z "$int_ip" ] && exit 1

ip netns add $router
#ip netns exec $router iptables -A INPUT -m mark --mark 0x1/0xffff -j ACCEPT
ip netns exec $router ip link set lo up
suffix=$1

./create_veth.sh $router ext-$suffix te-$suffix
ip netns exec $router ipset create nonat nethash
if [ -n "$ext_ip" ]; then
    eip=${ext_ip%/*}
    ip netns exec $router iptables -t nat -A POSTROUTING -m set --match-set nonat src -m set ! --match-set nonat dst -j SNAT --to-source $eip
fi
ip netns exec $router iptables -t nat -A POSTROUTING -m mark --mark 0x400 -j MASQUERADE
apply_vnic -I ext-$suffix

./create_veth.sh $router int-$suffix ti-$suffix
if [ -n "$int_ip" ]; then
    iip=${int_ip%/*}
    ip netns exec $router ipset add nonat $int_ip
    ip netns exec $router iptables -t nat -A POSTROUTING -d $int_ip -j SNAT --to-source $iip
fi
apply_vnic -I int-$suffix

router_dir=$cache_dir/router/$router
mkdir -p $router_dir
ip netns exec $router iptables-save > $router_dir/iptables.save
ip netns exec $router ipset save > $router_dir/ipset.save

vrrp_conf=$router_dir/keepalived.conf
notify_sh=$router_dir/notify.sh
cat > $vrrp_conf <<EOF
vrrp_instance $router {
    interface ns-${vrrp_vni}
    track_interface {
        ns-${vrrp_vni}
        ti-$suffix
        te-$suffix
    }
    dont_track_primary
    state $role
    virtual_router_id 100
    priority 100
    nopreempt
    advert_int 1

    virtual_ipaddress {
        $int_ip dev ti-$suffix
        $ext_ip dev te-$suffix
    }
    notify $notify_sh
}
EOF
cat > $notify_sh <<EOF
#!/bin/bash

TYPE=\$1
NAME=\$2
STATE=\$3

case \$STATE in
   "MASTER") 
        ip netns exec $router route add default gw $ext_gw
        ip netns exec $router arping -c 3 -I te-$suffix -s $eip $eip 
        ip netns exec $router arping -c 3 -I ti-$suffix -s $iip $iip
        exit 0
        ;;
   "BACKUP") 
        exit 0
        ;;
   "FAULT") 
        exit 0
        ;;
    *)  echo "unknown state"
        exit 1
    ;;
esac
EOF
chmod +x $notify_sh
./set_gateway.sh $router $vrrp_ip $vrrp_vni hard
pid_file=$router_dir/keepalived.pid
ip netns exec $router keepalived -f $vrrp_conf -p $pid_file -r $router_dir/vrrp.pid -c $router_dir/checkers.pid

interfaces=$(cat)
i=0
n=$(jq length <<< $interfaces)
while [ $i -lt $n ]; do
    addr=$(jq -r .[$i].ip_address <<< $interfaces)
    vni=$(jq -r .[$i].vni <<< $interfaces)
    routes=$(jq -r .[$i].routes <<< $interfaces)
    ./set_gw_route.sh $router $addr $vni soft <<< $routes
    let i=$i+1
done

ip netns exec $router bash -c "echo 1 >/proc/sys/net/ipv4/ip_forward"
echo "|:-COMMAND-:| $(basename $0) '$ID' '$SCI_CLIENT_ID' '$role'"
