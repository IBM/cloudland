#!/bin/bash

cland_root_dir=/opt/cloudland

# folders used by cloudland
mkdir -p $cland_root_dir/{etc,log,run,cache} $cland_root_dir/cache/{image,instance,dnsmasq,meta,router,volume,xml}

# download cloudland from github
function get_commitid()
{
    cd $cland_root_dir
    commitID=$(git rev-parse HEAD 2>/dev/null)
    if [ "$commitID" = "" ]; then
        commitID="not available"
    fi
    echo "$commitID" > commitID
}

# Build sci and Install it to /opt/sci
function inst_sci()
{
    cd $cland_root_dir/sci
    ./configure
    make
    make install
}

# Build and install cloudland
function inst_cland()
{
    # Setup grpc env
    export PATH=$PATH:/usr/local/bin
    export PKG_CONFIG_PATH=/usr/local/lib/pkgconfig

    # Build and install cland to /opt/cloudland/bin and /opt/cloudland/lib64
    cd $cland_root_dir/src
    make clean
    make
    make install

    # update cloudland.j2, the default SCI libraray path is sci/lib64
    if [ -d "/opt/sci/lib" ]; then
        echo "Update sci lib ..."
        sed -i "s/\/sci\/lib64/\/sci\/lib/g" $cland_root_dir/deploy/roles/cland/templates/cloudland.j2
        sed -i "s/\/sci\/lib64/\/sci\/lib/g" $cland_root_dir/deploy/roles/hyper/templates/cloudlet.j2
    fi
}

# Build web/clui
function build_clui()
{
    chown -R cland:cland $cland_root_dir

su cland << EOF
    # Prepare GO mod
    cd $cland_root_dir
    rm -f go.mod
    go mod init web
    go mod tidy
    echo 'replace github.com/IBM/cloudland => /opt/cloudland' >> go.mod

    # Build
    cd $cland_root_dir/web/clui
    go build
EOF

}

# Build libvirt-console-proxy
function inst_console_proxy()
{
    cd /opt
    git clone https://github.com/libvirt/libvirt-console-proxy.git
    chown -R cland:cland libvirt-console-proxy

su cland << EOF
    cd /opt/libvirt-console-proxy
    go build -o build/virtconsoleproxyd cmd/virtconsoleproxyd/virtconsoleproxyd.go
EOF
}

# Download noVNC
function get_noVNC()
{
    git clone https://github.com/novnc/noVNC.git $cland_root_dir/web/clui/public/novnc
    rm -rf $cland_root_dir/web/clui/public/novnc/.git*
    chown -R cland:cland $cland_root_dir
}

# create user cland if it is necessary
grep -E "cland" /etc/passwd > /dev/null 2>&1
if [ $? -ne 0 ]; then
    useradd cland
    echo 'cland ALL=(ALL) NOPASSWD:ALL' >> /etc/sudoers.d/cland
else
    echo "User cland already exists."
fi

# configure golang for cland
su cland << EOF
    # Prepare GO env
    sed -i '/export GO/d' ~/.bashrc
    echo 'export GOPROXY=https://goproxy.io' >> ~/.bashrc
    echo 'export GO111MODULE=on' >> ~/.bashrc
    source ~/.bashrc
EOF

get_commitid
inst_sci
inst_cland
build_clui
inst_console_proxy
get_noVNC