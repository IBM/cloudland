#!/bin/bash

[ $# -lt 3 ] && echo "$0 <cluster_name> <base_domain> <external_address>" && exit 1

cluster_name=$1
base_domain=$2
ext_addr=$3

seq_max=100
dns_server=$(grep '^namaserver' /etc/resolv.conf | tail -1 | awk '{print $2}')
if [ -n "$dns_server" -o "$dns_server" = "127.0.0.1" ]; then
    dns_server=8.8.8.8
fi
yum install -y dnsmasq

cp /etc/dnsmasq.conf /etc/dnsmasq.conf.bak

cat > /etc/dnsmasq.conf <<EOF
no-resolv
server=8.8.8.8
local=/${cluster_name}.${base_domain}/
address=/apps.${cluster_name}.${base_domain}/$ext_addr
srv-host=_etcd-server-ssl._tcp.${cluster_name}.${base_domain},etcd-0.${cluster_name}.${base_domain},2380,0,10
srv-host=_etcd-server-ssl._tcp.${cluster_name}.${base_domain},etcd-1.${cluster_name}.${base_domain},2380,0,10
srv-host=_etcd-server-ssl._tcp.${cluster_name}.${base_domain},etcd-2.${cluster_name}.${base_domain},2380,0,10
EOF
cat >> /etc/dnsmasq.conf <<EOF
no-hosts
addn-hosts=/etc/dnsmasq.openshift.addnhosts
conf-dir=/etc/dnsmasq.d,.rpmnew,.rpmsave,.rpmorig
EOF

cat > /etc/dnsmasq.openshift.addnhosts <<EOF
192.168.91.8 dns.${cluster_name}.${base_domain}
192.168.91.8 loadbalancer.${cluster_name}.${base_domain}  api.${cluster_name}.${base_domain}  api-int.${cluster_name}.${base_domain}
192.168.91.9 bootstrap.${cluster_name}.${base_domain}
192.168.91.10 master-0.${cluster_name}.${base_domain}  etcd-0.${cluster_name}.${base_domain}
192.168.91.11 master-1.${cluster_name}.${base_domain}  etcd-1.${cluster_name}.${base_domain}
192.168.91.12 master-2.${cluster_name}.${base_domain}  etcd-2.${cluster_name}.${base_domain}
EOF
for i in $(seq 0 $seq_max); do
    let suffix=$i+20
    cat >> /etc/dnsmasq.openshift.addnhosts <<EOF
192.168.91.$suffix worker-$i.${cluster_name}.${base_domain}
EOF
done

echo "nameserver 127.0.0.1" > /etc/resolv.conf
systemctl restart dnsmasq
systemctl enable dnsmasq
