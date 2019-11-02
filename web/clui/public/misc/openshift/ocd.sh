#!/bin/bash

cd $(dirname $0)

[ $# -lt 7 ] && echo "$0 <cluster_id> <cluster_name> <base_domain> <endpoint> <cookie> <ha_flag> <nworkers>" && exit 1

cluster_id=$1
cluster_name=$2
base_domain=$3
endpoint=$4
cookie=$5
haflag=$6
nworkers=$7
seq_max=100

function setup_dns()
{
    instID=$(cat /var/lib/cloud/data/instance-id | cut -d'-' -f2)
    count=0
    while [ -z "$public_ip" -a $count -lt 10 ]; do
        data=$(curl -k -XPOST $endpoint/floatingips/assign --cookie "$cookie" --data "instance=$instID")
        echo $data
        public_ip=$(jq  -r .public_ip <<< $data)
        public_ip=${public_ip%%/*}
        let count=$count+1
        sleep 1
    done
    [ -z "$public_ip" ] && public_ip=192.168.91.8
    dns_server=$(grep '^nameserver' /etc/resolv.conf | head -1 | awk '{print $2}')
    if [ -z "$dns_server" -o "$dns_server" = "127.0.0.1" ]; then
        dns_server=8.8.8.8
    fi

    yum install -y dnsmasq
    cp /etc/dnsmasq.conf /etc/dnsmasq.conf.bak
    cat > /etc/dnsmasq.conf <<EOF
no-resolv
server=$dns_server
local=/${cluster_name}.${base_domain}/
address=/apps.${cluster_name}.${base_domain}/$public_ip
srv-host=_etcd-server-ssl._tcp.${cluster_name}.${base_domain},etcd-0.${cluster_name}.${base_domain},2380,0,10
EOF
    if [ "$haflag" = "yes" ]; then
        cat >> /etc/dnsmasq.conf <<EOF
srv-host=_etcd-server-ssl._tcp.${cluster_name}.${base_domain},etcd-1.${cluster_name}.${base_domain},2380,0,10
srv-host=_etcd-server-ssl._tcp.${cluster_name}.${base_domain},etcd-2.${cluster_name}.${base_domain},2380,0,10
EOF
    fi
    cat >> /etc/dnsmasq.conf <<EOF
no-hosts
addn-hosts=/etc/dnsmasq.openshift.addnhosts
conf-dir=/etc/dnsmasq.d,.rpmnew,.rpmsave,.rpmorig
EOF

    cat > /etc/dnsmasq.openshift.addnhosts <<EOF
$public_ip dns.${cluster_name}.${base_domain}
$public_ip loadbalancer.${cluster_name}.${base_domain}  api.${cluster_name}.${base_domain}  api-int.${cluster_name}.${base_domain}  lb.${cluster_name}.${base_domain}
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
    yum install -y wget nc jq
    cd /opt
    wget --no-check-certificate $endpoint/misc/openshift/ocd.conf
    source ocd.conf
    wget --no-check-certificate -O /usr/share/nginx/html/rhcos.raw.gz $coreos_image_url
    wget --no-check-certificate -O openshift-install-linux.tgz $openshift_installer
    wget --no-check-certificate -O openshift-client-linux.tgz $openshift_client
    tar -zxf openshift-install-linux.tgz
    tar -zxf openshift-client-linux.tgz
    cp kubectl oc /usr/bin/
}

function ignite_files()
{
    secret=$(cat)
    ssh_key=$(cat /home/centos/.ssh/authorized_keys | tail -1)
    rm -rf $cluster_name
    mkdir $cluster_name
    cd $cluster_name
    mreplica=1
    [ "$haflag" = "yes" ] && mreplica=3
    cat > install-config.yaml <<EOF
apiVersion: v1
baseDomain: $base_domain
compute:
- hyperthreading: Enabled
  name: worker
  platform: {}
  replicas: 0
controlPlane:
  hyperthreading: Enabled
  name: master
  platform: {}
  replicas: $mreplica
metadata:
  name: $cluster_name
networking:
  clusterNetwork:
  - cidr: 10.128.0.0/14
    hostPrefix: 23
  networkType: OpenShiftSDN
  serviceNetwork:
  - 172.30.0.0/16
platform:
  none: {}
pullSecret: '$secret'
sshKey: '$ssh_key'
EOF
    ../openshift-install create manifests
    sed -i "s/mastersSchedulable: true/mastersSchedulable: false/" manifests/cluster-scheduler-02-config.yml
    ../openshift-install create ignition-configs
    ignite_dir=/usr/share/nginx/html/ignition
    rm -rf $ignite_dir
    mkdir $ignite_dir
    cp *.ign $ignite_dir
    cat >>/root/.bashrc <<EOF
export KUBECONFIG=/opt/$cluster_name/auth/kubeconfig
export PS1='[\u@\h.$cluster_name \w]\\$ '
EOF
    cat >>/home/centos/.bashrc <<EOF
export PS1='[\u@\h.$cluster_name \w]\\$ '
EOF
}

function setup_nfs_pv()
{
    cd /opt/$cluster_name
    mkdir data
    yum -y install nfs-utils nfs-utils-lib
    systemctl start  rpcbind
    systemctl start nfs
    systemctl start  nfslock
    systemctl enable rpcbind
    systemctl enable nfs
    systemctl enable  nfslock
    cat >/etc/exports <<EOF
/opt/$cluster_name/data 192.168.91.0/24(rw,sync,no_root_squash,no_subtree_check,insecure)
EOF
    exportfs -a

    cat >nfs-pv.yaml <<EOF
apiVersion: v1
kind: PersistentVolume
metadata:
  name: nfs-pv
spec:
  capacity:
    storage: 100Gi 
  accessModes:
  - ReadWriteMany 
  nfs: 
    path: /opt/$cluster_name/data 
    server: 192.168.91.8
  persistentVolumeReclaimPolicy: Recycle
EOF
    ../oc create -f nfs-pv.yaml
    ../oc patch configs.imageregistry/cluster --type merge --patch '{"spec":{"storage":{"pvc":{"claim":""}}}}'
    cd -
}

function create_storage()
{
    cd /opt/$cluster_name
    cat >storage.yaml <<EOF
kind: StorageClass
apiVersion: storage.k8s.io/v1
metadata:
  name: gp2
provisioner: kubernetes.io/glusterfs
parameters:
  resturl: 'http://192.168.91.199:8080'
reclaimPolicy: Delete
volumeBindingMode: Immediate
EOF
    ../oc create -f storage.yaml
    cd -
}

function set_autocsr()
{
    cat >/etc/cron.hourly/50autocsr <<EOF
#!/bin/bash
export KUBECONFIG=/opt/$cluster_name/auth/kubeconfig
/opt/oc get csr -ojson | jq -r '.items[] | select(.status == {} ) | .metadata.name' | xargs /opt/oc adm certificate approve
EOF
    chmod +x /etc/cron.hourly/50autocsr
}

function launch_cluster()
{
    cd /opt/$cluster_name
    bstrap_res=$(curl -k -XPOST $endpoint/openshifts/$cluster_id/launch --cookie $cookie --data "hostname=bootstrap;ipaddr=192.168.91.9")
    bstrap_ID=$(jq -r .ID <<< $bstrap_res)
    curl -k -XPOST $endpoint/openshifts/$cluster_id/state --cookie $cookie --data "status=bootstrap"
    while true; do
        sleep 5
        nc -zv 192.168.91.9 6443
        [ $? -eq 0 ] && break
    done
    curl -k -XPOST $endpoint/openshifts/$cluster_id/state --cookie $cookie --data "status=masters"
    curl -k -XPOST $endpoint/openshifts/$cluster_id/launch --cookie $cookie --data "hostname=master-0;ipaddr=192.168.91.10"
    sleep 3
    if [ "$haflag" = "yes" ]; then
        curl -k -XPOST $endpoint/openshifts/$cluster_id/launch --cookie $cookie --data "hostname=master-1;ipaddr=192.168.91.11"
        sleep 3
        curl -k -XPOST $endpoint/openshifts/$cluster_id/launch --cookie $cookie --data "hostname=master-2;ipaddr=192.168.91.12"
        sleep 3
    fi
    curl -k -XPOST $endpoint/openshifts/$cluster_id/state --cookie $cookie --data "status=workers"
    curl -k -XPOST $endpoint/openshifts/$cluster_id/launch --cookie $cookie --data "hostname=worker-0;ipaddr=192.168.91.20"
    sleep 3
    curl -k -XPOST $endpoint/openshifts/$cluster_id/launch --cookie $cookie --data "hostname=worker-1;ipaddr=192.168.91.21"
    sleep 3
    sleep 60
    ../openshift-install wait-for bootstrap-complete --log-level debug
    curl -k -XDELETE $endpoint/instances/$bstrap_ID --cookie $cookie
    nodes=3
    [ "$haflag" = "yes" ] && nodes=5
    export KUBECONFIG=auth/kubeconfig
    while true; do
        ../oc get csr -ojson | jq -r '.items[] | select(.status == {} ) | .metadata.name' | xargs ../oc adm certificate approve
        sleep 5
        count=$(../oc get nodes | grep -c Ready)
        [ "$count" -ge "$nodes" ] && break
    done
    sleep 60
    while true; do
        sleep 5
        ../oc get csr -ojson | jq -r '.items[] | select(.status == {} ) | .metadata.name' | xargs ../oc adm certificate approve
        ../oc get clusteroperators image-registry
        [ $? -eq 0 ] && break
    done
    setup_nfs_pv
    ../openshift-install wait-for install-complete
    curl -k -XPOST $endpoint/openshifts/$cluster_id/state --cookie $cookie --data "status=complete"
    let more=$nworkers-2
    for i in $(seq 1 $more); do
        let index=$i+1
        let last=$index+20
        curl -k -XPOST $endpoint/openshifts/$cluster_id/launch --cookie $cookie --data "hostname=worker-$index;ipaddr=192.168.91.$last"
        sleep 3
    done
    let nodes=$nodes+$more
    while true; do
        sleep 5
        ../oc get csr -ojson | jq -r '.items[] | select(.status == {} ) | .metadata.name' | xargs ../oc adm certificate approve
        count=$(../oc get nodes | grep -c Ready)
        [ "$count" -ge "$nodes" ] && break
    done
    set_autocsr
    # create_storage
}

setenforce Permissive
sed -i 's/^SELINUX=enforcing/SELINUX=permissive/' /etc/selinux/config
yum -y install epel-release
yum -y install jq
setup_dns
setup_lb
setup_nginx
download_pkgs
ignite_files
launch_cluster
