#!/bin/bash

datetime=$(date +'%Y-%m-%d-%H:%M:%S')
logfile=/tmp/allinone-deploy-$datetime.log
echo "Install is in progress... Log file is $logfile"
exec &> >(tee $logfile)

cland_root_dir=/opt/cloudland
cd $(dirname $0)
[ $PWD != "$cland_root_dir/deploy" ] && echo "Please clone cloudland into /opt" && exit 1
net_conf=$cland_root_dir/deploy/netconf.yml

sudo chown -R cland.cland $cland_root_dir
mkdir $cland_root_dir/{bin,deploy,etc,lib6,log,run,sci,scripts,src,web,cache} $cland_root_dir/cache/{image,instance,dnsmasq,meta,router,volume,xml} 2>/dev/null
[ ! -s "$net_conf" ] && sudo cp ${net_conf}.example $net_conf

# Install development tools
[ $(uname -m) != s390x ] && sudo yum -y install epel-release
sudo yum install -y ansible vim git wget net-tools
sudo yum groupinstall -y "Development Tools"

# Install SCI
function inst_sci() 
{
    cd $cland_root_dir/sci
    ./configure
    make
    sudo make install
}

# Install GRPC
function inst_grpc() {
    sudo yum install -y axel
    cd $cland_root_dir
    grpc_pkg=/tmp/grpc.tar.gz
    grpc_url='http://www.bluecat.ltd/repo/grpc.tar.gz'
    grep -q 'release 8' /etc/redhat-release
    [ $? -eq 0 ] && grpc_url='http://www.bluecat.ltd/repo/grpc8.tar.gz'
    wget $grpc_url -O $grpc_pkg
    sudo tar -zxf $grpc_pkg -C /
    rm -f $grpc_pkg
    sudo bash -c 'echo /usr/local/lib > /etc/ld.so.conf.d/protobuf.conf'
    sudo ldconfig
}

# Install web
function inst_web()
{
    cd $cland_root_dir/deploy
    ansible-playbook cloudland.yml -e @$net_conf --tags database
    sudo yum -y install golang
    sudo chown -R cland.cland /usr/local
    sed -i '/export GO/d' ~/.bashrc
    echo 'export GOPROXY=https://goproxy.io' >> ~/.bashrc
    echo 'export GO111MODULE=on' >> ~/.bashrc
    source ~/.bashrc
    cd $cland_root_dir
    rm -f go.mod
    go mod init web
    go mod tidy
    echo 'replace github.com/IBM/cloudland => /opt/cloudland' >> go.mod
    cd $cland_root_dir/web/clui
    go build
    cd $cland_root_dir/deploy
    ansible-playbook cloudland.yml -e @$net_conf --tags web
}

# Install cloudland
function inst_cland()
{
    cd $cland_root_dir/src
    export PKG_CONFIG_PATH=/usr/local/lib/pkgconfig
    make clean
    make
    make install
}

# Install libvirt console proxy
function inst_console_proxy()
{
    sudo yum -y install libvirt-devel
    cd /opt
    sudo git clone https://github.com/libvirt/libvirt-console-proxy.git
    sudo chown cland.cland libvirt-console-proxy
    cd libvirt-console-proxy
    go build -o build/virtconsoleproxyd cmd/virtconsoleproxyd/virtconsoleproxyd.go
    git clone https://github.com/novnc/noVNC.git /opt/cloudland/web/clui/public/novnc
    rm -rf /opt/cloudland/web/clui/public/novnc/.git*
    cd $cland_root_dir/deploy
}

# Generate host file
function gen_hosts()
{
    cland_ssh_dir=$cland_root_dir/deploy/.ssh
    mkdir -p $cland_ssh_dir
    chmod 700 $cland_ssh_dir
    if [ ! -f $cland_ssh_dir/cland.key ]; then
        yes y | ssh-keygen -t rsa -N "" -f $cland_ssh_dir/cland.key
        mkdir -p ~/.ssh
        chmod 700 ~/.ssh
        touch ~/.ssh/authorized_keys
        chmod 600 ~/.ssh/authorized_keys
        cat $cland_ssh_dir/cland.key.pub >> ~/.ssh/authorized_keys
    fi

    net_dev=$(cat $net_conf | grep 'network_device:' | cut -d: -f2)
    myip=$(ifconfig $net_dev | grep 'inet ' | awk '{print $2}')
    hname=$(hostname -s)
    sudo bash -c "echo '$myip $hname' >> /etc/hosts"
    echo $hname > $cland_root_dir/etc/host.list
    mkdir -p $cland_root_dir/deploy/hosts
    virt_type=kvm-x86_64
    [ "$hyper_type" != "x86_64" ] && virt_type=kvm-s390x
    cat > $cland_root_dir/deploy/hosts/hosts <<EOF
[hyper]
$hname ansible_host=$myip ansible_ssh_private_key_file=$cland_ssh_dir/cland.key client_id=0 zone_name=zone0 virt_type=$virt_type

[cland]
$hname ansible_host=$myip ansible_ssh_private_key_file=$cland_ssh_dir/cland.key

[web]
$hname ansible_host=$myip ansible_ssh_private_key_file=$cland_ssh_dir/cland.key

[database]
$hname ansible_host=$myip ansible_ssh_private_key_file=$cland_ssh_dir/cland.key
EOF
}

function demo_router()
{
    ext_vlan=$(cat $net_conf | grep 'network_external_vlan:' | cut -d: -f2 | xargs)
    int_vlan=$(cat $net_conf | grep 'network_internal_vlan:' | cut -d: -f2 | xargs)
    br_ext=br$ext_vlan
    br_int=br$int_vlan
    sudo /opt/cloudland/scripts/backend/create_link.sh $ext_vlan
    sudo /opt/cloudland/scripts/backend/create_link.sh $int_vlan
    sudo nmcli connection modify $br_ext ipv4.addresses 192.168.71.1/24
    sudo nmcli connection modify $br_int ipv4.addresses 172.16.20.1/24
    sudo nmcli connection up $br_ext
    sudo nmcli connection up $br_int
    sudo grep -q "^GatewayPorts yes" /etc/ssh/sshd_config
    [ $? -ne 0 ] && sudo bash -c "echo -e '\nGatewayPorts yes' >> /etc/ssh/sshd_config"
    sudo systemctl restart sshd
}

function allinone_firewall()
{
    sudo iptables -D INPUT -p tcp -m state --state NEW -m tcp --dport 80 -j ACCEPT
    sudo iptables -I INPUT -p tcp -m state --state NEW -m tcp --dport 80 -j ACCEPT
    sudo iptables -D INPUT -p tcp -m state --state NEW -m tcp --dport 443 -j ACCEPT
    sudo iptables -I INPUT -p tcp -m state --state NEW -m tcp --dport 443 -j ACCEPT
    sudo iptables -D INPUT -p tcp -m state --state NEW -m tcp --dport 4000 -j ACCEPT
    sudo iptables -I INPUT -p tcp -m state --state NEW -m tcp --dport 4000 -j ACCEPT
    sudo iptables -D INPUT -p tcp -m state --state NEW -m tcp --dport 9988 -j ACCEPT
    sudo iptables -I INPUT -p tcp -m state --state NEW -m tcp --dport 9988 -j ACCEPT
    sudo iptables -D INPUT -p tcp -m state --state NEW -m tcp --dport 18000:20000 -j ACCEPT
    sudo iptables -I INPUT -p tcp -m state --state NEW -m tcp --dport 18000:20000 -j ACCEPT
    sudo service iptables save
}

export PATH=$PATH:/usr/local/bin
diff /opt/sci/lib64/libsci.so.0.0.0 $cland_root_dir/sci/libsci/.libs/libsci.so.0.0.0
[ $? -ne 0 ] && inst_sci
[ ! -d "/usr/local/lib/pkgconfig" ] && inst_grpc
diff $cland_root_dir/bin/cloudland $cland_root_dir/src/cloudland
[ $? -ne 0 ] && inst_cland

hyper_type=$(uname -m)
gen_hosts
cd $cland_root_dir/deploy
[ "$hyper_type" != s390x ] && ansible-playbook cloudland.yml -e @$net_conf --tags epel
ansible-playbook cloudland.yml -e @$net_conf --tags hosts,selinux,be_pkg,be_conf,firewall
allinone_firewall
inst_web
inst_console_proxy
ansible-playbook cloudland.yml -e @$net_conf --tags be_srv,fe_srv,console,imgrepo
demo_router
sudo chown -R cland.cland $cland_root_dir

echo "Installation completes. Log file is /tmp/allinone-deploy-2021-01-10-10:42:09.log"
