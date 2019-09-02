#!/bin/bash

echo "$@" > /tmp/openshift
exit 0

cd $(dirname $0)

[ $# -lt 5 ] && echo "$0 <cluster_name> <base_domain> <endpoint> <cookie> <ha_flag>"

cluster_name=$1
base_domain=$2
endpoint=$3
cookie=$4
haflag=$5
seq_max=100

function setup_dns()
{
    instID=$(cat /var/lib/cloud/data/instance-id | cut -d'-' -f2)
    data=$(curl -XPOST $endpoint/floatingips/assign --cookie "$cookie" --form "instance=$instID")
    public_ip=$(jq  -r .public_ip <<< $data)
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
address=/apps.${cluster_name}.${base_domain}/$public_ip
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
}

function setup_lb()
{
    yum install -y haproxy
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
}

function setup_nginx()
{
    yum install -y epel-release
    yum install -y nginx
    cp /etc/nginx/nginx.conf /etc/nginx/nginx.conf.bak
    cat > /etc/nginx/nginx.conf <<EOF
user nginx;
worker_processes auto;
error_log /var/log/nginx/error.log;
pid /run/nginx.pid;

# Load dynamic modules. See /usr/share/nginx/README.dynamic.
include /usr/share/nginx/modules/*.conf;

events {
    worker_connections 1024;
}

http {
    log_format  main  '\$remote_addr - \$remote_user [\$time_local] "\$request" '
                      '\$status \$body_bytes_sent "\$http_referer" '
                      '"\$http_user_agent" "\$http_x_forwarded_for"';

    access_log  /var/log/nginx/access.log  main;

    sendfile            on;
    tcp_nopush          on;
    tcp_nodelay         on;
    keepalive_timeout   65;
    types_hash_max_size 2048;

    include             /etc/nginx/mime.types;
    default_type        application/octet-stream;

    # Load modular configuration files from the /etc/nginx/conf.d directory.
    # See http://nginx.org/en/docs/ngx_core_module.html#include
    # for more information.
    include /etc/nginx/conf.d/*.conf;

    server {
        listen       8080 default_server;
        listen       [::]:8080 default_server;
        server_name  _;
        root         /usr/share/nginx/html;

        # Load configuration files for the default server block.
        include /etc/nginx/default.d/*.conf;

        location / {
        }

        error_page 404 /404.html;
            location = /40x.html {
        }

        error_page 500 502 503 504 /50x.html;
            location = /50x.html {
        }
    }
}
EOF

    systemctl restart nginx
    systemctl enable nginx
}

function download_pkgs()
{
    yum install -y wget
    wget -O /usr/share/nginx/html/rhcos.raw.gz https://mirror.openshift.com/pub/openshift-v4/dependencies/rhcos/4.1/latest/rhcos-4.1.0-x86_64-metal-bios.raw.gz
    cd /opt
    wget https://mirror.openshift.com/pub/openshift-v4/clients/ocp/latest/openshift-client-linux-4.1.11.tar.gz
    wget https://mirror.openshift.com/pub/openshift-v4/clients/ocp/latest/openshift-install-linux-4.1.11.tar.gz
    tar -zxf openshift-client-linux-4.1.11.tar.gz
    tar -zxf openshift-install-linux-4.1.11.tar.gz
}

setenforce Permissive
sed -i 's/^SELINUX=enforcing/SELINUX=permissive/' /etc/selinux/config
setup_dns
setup_lb
setup_nginx
download_pkgs
