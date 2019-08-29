#!/bin/bash

[ $# -lt 2 ] && echo "$0 <cluster_name> <base_domain>"

cluster_name=$1
base_domain=$2

seq_max=100
yum install -y haproxy
setenforce Permissive
sed -i 's/^SELINUX=enforcing/SELINUX=permissive/' /etc/selinux/config

haconf=/etc/haproxy/haproxy.cfg
cp $haconf ${haconf}.bak
cat > $haconf <<EOF
global
    log         127.0.0.1 local2 info
    chroot      /var/lib/haproxy
    pidfile     /var/run/haproxy.pid
    maxconn     4000
    user        haproxy
    group       haproxy
    daemon

defaults
    timeout connect         5s
    timeout client          30s
    timeout server          30s
    log                     global

frontend kubernetes_api
    bind 0.0.0.0:6443
    default_backend kubernetes_api

frontend machine_config
    bind 0.0.0.0:22623
    default_backend machine_config

frontend router_https
    bind 0.0.0.0:443
    default_backend router_https

frontend router_http
    mode http
    option httplog
    bind 0.0.0.0:80
    default_backend router_http

backend kubernetes_api
    balance roundrobin
    option ssl-hello-chk
    server bootstrap bootstrap.${cluster_name}.${base_domain}:6443 check
    server master-0 master-0.${cluster_name}.${base_domain}:6443 check
    server master-1 master-1.${cluster_name}.${base_domain}:6443 check
    server master-2 master-2.${cluster_name}.${base_domain}:6443 check

backend machine_config
    balance roundrobin
    option ssl-hello-chk
    server bootstrap bootstrap.${cluster_name}.${base_domain}:22623 check
    server master-0 master-0.${cluster_name}.${base_domain}:22623 check
    server master-1 master-1.${cluster_name}.${base_domain}:22623 check
    server master-2 master-2.${cluster_name}.${base_domain}:22623 check

backend router_https
    balance roundrobin
    option ssl-hello-chk
EOF

for i in $(seq 0 $seq_max); do
    cat >> $haconf <<EOF
    server worker-$i worker-$i.${cluster_name}.${base_domain}:443 check
EOF
done
cat >> $haconf <<EOF

backend router_http
    mode http
    balance roundrobin
EOF
for i in $(seq 0 $seq_max); do
    cat >> $haconf <<EOF
    server worker-$i worker-$i.${cluster_name}.${base_domain}:80 check
EOF
done

systemctl restart haproxy
systemctl enable haproxy
