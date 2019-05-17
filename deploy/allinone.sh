#!/bin/bash

DB_PASSWD=passw0rd
NET_DEV=eth0

cland_root_dir=/opt/cloudland
cd $(dirname $0)

[ $PWD != "$cland_root_dir/deploy" ] && echo "Please clone cloudland into /opt" && exit 1

sudo chown -R centos.centos /opt/cloudland

# Install development tools
sudo yum install -y ansible vim git wget epel-release
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
    cd $cland_root_dir
    git clone https://github.com/grpc/grpc.git
    cd grpc
    git submodule update --init
    cd third_party/protobuf/
    ./autogen.sh
    ./configure
    make
    sudo make install
    sudo bash -c 'echo /usr/local/lib > /etc/ld.so.conf.d/protobuf.conf'
    sudo ldconfig
    cd ../../
    make
    sudo make install
    DIR=$PWD
    cat > activate.sh <<EOF
export PATH=$PATH:/usr/local/bin:$DIR/bins/opt:$DIR/bins/opt/protobuf
export CPATH=$DIR/include:$DIR/third_party/protobuf/src
export LIBRARY_PATH=$DIR/libs/opt:$DIR/libs/opt/protobuf
export PKG_CONFIG_PATH=$DIR/libs/opt/pkgconfig:$DIR/third_party/protobuf
export LD_LIBRARY_PATH=$DIR/libs/opt
EOF
}

# Install web
function inst_web()
{
    cd $cland_root_dir/deploy
    ansible-playbook cloudland.yml --tags database --extra-vars "db_passwd=$DB_PASSWD"
    sudo yum -y install golang 
    sudo chown -R centos.centos /usr/local
    sed -i '/export GO/d' ~/.bashrc
    echo 'export GOPATH=/usr/local' >> ~/.bashrc
    echo 'export GOPROXY=https://goproxy.io' >> ~/.bashrc
    source ~/.bashrc
    go get github.com/IBM/cloudland
    cd /usr/local/src/github.com/IBM/cloudland/web/clui
    echo 'export GO111MODULE=on' >> ~/.bashrc
    source ~/.bashrc
    go build
    cd $cland_root_dir/deploy
    ansible-playbook cloudland.yml --tags web --extra-vars "db_passwd=$DB_PASSWD"
}

# Install cloudland
function inst_cland()
{
    cd $cland_root_dir/src
    source $cland_root_dir/grpc/activate.sh
    make
    make install
}

# Generate host file
function gen_hosts()
{
    myip=$(ifconfig $NET_DEV | grep 'inet ' | awk '{print $2}')
    sudo bash -c "sed -i '/$myip $hname/d' /etc/hosts"
    hname=$(hostname -s)
    sudo bash -c "echo '$myip $hname' >> /etc/hosts"
    mkdir $cland_root_dir/{etc,log}
    echo $hname > $cland_root_dir/etc/host.list
    cat > $cland_root_dir/deploy/hosts/hosts <<EOF
[hyper]
$hname ansible_host=$myip client_id=0

[cland]
$hname ansible_host=$myip

[web]
$hname ansible_host=$myip

[database]
$hname ansible_host=$myip
EOF
}

diff /opt/sci/lib64/libsci.so.0.0.0 $cland_root_dir/sci/libsci/.libs/libsci.so.0.0.0
[ $? -ne 0 ] && inst_sci
[ ! -f "$cland_root_dir/grpc/activate.sh" ] && inst_grpc
diff $cland_root_dir/bin/cloudland $cland_root_dir/src/cloudland
[ $? -ne 0 ] && inst_cland

gen_hosts
ansible-playbook cloudland.yml --tags be_srv,fe_srv
inst_web
