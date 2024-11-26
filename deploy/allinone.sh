#!/bin/bash

datetime=$(date +'%Y-%m-%d-%H:%M:%S')
logfile=/tmp/allinone-deploy-$datetime.log
echo "Install is in progress... Log file is $logfile"
exec &> >(tee $logfile)

cland_root_dir=/opt/cloudland
cd $(dirname $0)
[ $PWD != "$cland_root_dir/deploy" ] && echo "Please clone cloudland into /opt" && exit 1
net_conf=$cland_root_dir/deploy/netconf.yml

useradd -m -s /bin/bash cland
echo 'cland ALL=(ALL) NOPASSWD:ALL' > /etc/sudoers.d/cland
sudo chown -R cland.cland $cland_root_dir
mkdir $cland_root_dir/{bin,deploy,etc,lib6,log,run,sci,scripts,src,web,cache} $cland_root_dir/cache/{image,instance,dnsmasq,meta,router,volume,xml} 2>/dev/null
[ ! -s "$net_conf" ] && sudo cp ${net_conf}.example $net_conf

# Install packages
apt install -y make g++ libssl-dev libjsoncpp-dev ansible jq wget mkisofs network-manager net-tools python3-pip qemu-system-x86 qemu-utils bridge-utils ipcalc ipset iputils-arping libvirt-daemon libvirt-daemon-system libvirt-daemon-system-systemd libvirt-clients dnsmasq keepalived dnsmasq-utils conntrack
go version
if [ $? -ne 0 ]; then
    cd /tmp
    wget https://go.dev/dl/go1.21.13.linux-amd64.tar.gz
    tar -zxvf go1.21.13.linux-amd64.tar.gz
    mv go /usr/local/go
    export PATH=$PATH:/usr/local/go/bin:/root/go/bin
    echo 'PATH=$PATH:/usr/local/go/bin:/root/go/bin' >> $HOME/.bashrc
    go install github.com/swaggo/swag/cmd/swag@latest
    cd -
fi

# Install SCI
function inst_sci() 
{
    cd $cland_root_dir/sci
    ./configure
    make
    sudo make install
}

# Install web
function inst_web()
{
    cd $cland_root_dir/deploy
    ansible-playbook cloudland.yml -e @$net_conf --tags database
    cd $cland_root_dir/web
    make all
    cd $cland_root_dir/deploy
    ansible-playbook cloudland.yml -e @$net_conf --tags web
}

# Install cloudland
function inst_cland()
{
    cd $cland_root_dir/src
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

diff /opt/sci/lib64/libsci.so.0.0.0 $cland_root_dir/sci/libsci/.libs/libsci.so.0.0.0
[ $? -ne 0 ] && inst_sci
diff $cland_root_dir/bin/cloudland $cland_root_dir/src/cloudland
[ $? -ne 0 ] && inst_cland

gen_hosts
cd $cland_root_dir/deploy
ansible-playbook cloudland.yml -e @$net_conf --skip-tags fe_bin,sci,sync,firewall #--tags hosts,be_pkg,be_conf,be_srv
#allinone_firewall
inst_web
#inst_console_proxy
#ansible-playbook cloudland.yml -e @$net_conf --tags fe_srv,console,nginx
#sudo chown -R cland.cland $cland_root_dir

echo "Installation completes. Log file is /tmp/allinone-deploy-2021-01-10-10:42:09.log"
