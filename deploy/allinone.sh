#!/bin/bash

ADMIN_PASSWD=passw0rd
NET_DEV=eth0
DB_PASSWD=d6Passwd

cland_root_dir=/opt/cloudland
cd $(dirname $0)

[ $PWD != "$cland_root_dir/deploy" ] && echo "Please clone cloudland into /opt" && exit 1

sudo chown -R cland.cland $cland_root_dir
mkdir $cland_root_dir/{bin,deploy,etc,lib6,log,run,sci,scripts,src,web,cache} $cland_root_dir/cache/{image,instance,meta,router,volume,xml} 2>/dev/null

# Install development tools
sudo yum install -y ansible vim git wget epel-release net-tools
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
    axel -q -n 20 http://www.bluecat.ltd/repo/grpc.tar.gz -o $grpc_pkg
    sudo tar -zxf $grpc_pkg -C /
    rm -f $grpc_pkg
    sudo bash -c 'echo /usr/local/lib > /etc/ld.so.conf.d/protobuf.conf'
    sudo ldconfig
}

# Install web
function inst_web()
{
    cd $cland_root_dir/deploy
    ansible-playbook cloudland.yml --tags database --extra-vars "db_passwd=$DB_PASSWD"
    sudo yum -y install golang 
    sudo chown -R cland.cland /usr/local
    sed -i '/export GO/d' ~/.bashrc
    echo 'export GOPROXY=https://goproxy.io' >> ~/.bashrc
    echo 'export GO111MODULE=on' >> ~/.bashrc
    source ~/.bashrc
    cd $cland_root_dir/web/clui
    go build
    cd $cland_root_dir/deploy
    ansible-playbook cloudland.yml --tags web --extra-vars "db_passwd=$DB_PASSWD" --extra-vars "admin_passwd=$ADMIN_PASSWD"
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

    myip=$(ifconfig $NET_DEV | grep 'inet ' | awk '{print $2}')
    hname=$(hostname -s)
    sudo bash -c "echo '$myip $hname' >> /etc/hosts"
    echo $hname > $cland_root_dir/etc/host.list
    mkdir -p $cland_root_dir/deploy/hosts
    cat > $cland_root_dir/deploy/hosts/hosts <<EOF
[hyper]
$hname ansible_host=$myip ansible_ssh_private_key_file=$cland_ssh_dir/cland.key client_id=0

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
    sudo /opt/cloudland/scripts/backend/create_link.sh 5000
    sudo /opt/cloudland/scripts/backend/create_link.sh 5010
    sudo nmcli connection modify br5000 ipv4.addresses 192.168.71.1/24
    sudo nmcli connection modify br5010 ipv4.addresses 172.16.20.1/24
    sudo nmcli connection up br5000
    sudo nmcli connection up br5010
    sudo grep -q "^GatewayPorts yes" /etc/ssh/sshd_config
    [ $? -ne 0 ] && sudo bash -c "echo -e '\nGatewayPorts yes' >> /etc/ssh/sshd_config"
    sudo systemctl restart sshd
}

function allinone_firewall()
{
    sudo iptables -D INPUT -p tcp -m state --state NEW -m tcp --dport 80 -j ACCEPT
    sudo iptables -I INPUT -p tcp -m state --state NEW -m tcp --dport 80 -j ACCEPT
    sudo iptables -D INPUT -p tcp -m state --state NEW -m tcp --dport 4000 -j ACCEPT
    sudo iptables -I INPUT -p tcp -m state --state NEW -m tcp --dport 4000 -j ACCEPT
    sudo iptables -D INPUT -p tcp -m state --state NEW -m tcp --dport 18000:20000 -j ACCEPT
    sudo iptables -I INPUT -p tcp -m state --state NEW -m tcp --dport 18000:20000 -j ACCEPT
    sudo service iptables save
}

diff /opt/sci/lib64/libsci.so.0.0.0 $cland_root_dir/sci/libsci/.libs/libsci.so.0.0.0
[ $? -ne 0 ] && inst_sci
[ ! -f "/usr/local/lib/pkgconfig" ] && inst_grpc
diff $cland_root_dir/bin/cloudland $cland_root_dir/src/cloudland
[ $? -ne 0 ] && inst_cland

gen_hosts
cd $cland_root_dir/deploy
ansible-playbook cloudland.yml --tags hosts,epel,ntp,be_pkg,be_conf,be_srv,firewall,fe_srv,imgrepo --extra-vars "network_device=$NET_DEV"
inst_web
demo_router
allinone_firewall
