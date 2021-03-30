#!/bin/bash

# Check root
if [[ `whoami` != "root" ]]; then
    echo "Not root"
    exit -1
fi

if [ $# -lt 2 ]; then
    echo "$0 <version> <release>"
    exit -1
fi

arch=$(uname -m)
version=$1
release=$2

echo "Prepare packages..."

# Install rpmbuild
yum groupinstall -y "RPM Development Tools"

# Install tools
yum install -y rsync

# prepare grpc
commitID=$(cat /root/cloudland-grpc/commit)
if [[ "$commitID" = "" ]]; then
    echo "No grpc found from /root/cloudland-grpc. Refer to build_grpc.sh to generate the package."
    exit -1
fi
cp /root/cloudland-grpc/grpc-${commitID}.tar.gz /tmp/grpc.tar.gz

# prepare cloudland package
rm -rf /tmp/opt
mkdir -p /tmp/opt
# copy files to /tmp/opt
rsync -a --exclude={'.git','cache/*','run/*','log/*','etc/*','scripts/cloudrc.local','web/clui/conf/config.toml','web/clui/public/misc/openshift/*','deploy/conf.json'} --include={'web/clui/public/misc/openshift/ocd.sh'} /opt/cloudland /tmp/opt
rsync -a /opt/sci /tmp/opt
rsync -a --exclude={'cert/*'} /opt/libvirt-console-proxy /tmp/opt
cland_root_dir=/tmp/opt/cloudland
# clear cloudland git files
cd $cland_root_dir
commitID="unknown"
if [ -e "commitID" ]; then
    commitID=$(cat commitID)
fi
# clear sci
cd $cland_root_dir/sci
make clean > /dev/null 2>&1
# clear cloudland
cd $cland_root_dir/src
make clean > /dev/null 2>&1
# package cloudland
cd /tmp
tar -czf cloudland.tar.gz opt
rm -rf /tmp/opt

# do rpmbuild
cd ~
rm -rf rpmbuild
mkdir -p rpmbuild/{BUILD,BUILDROOT,RPMS,SOURCES,SPECS,SRPMS}
mkdir -p rpmbuild/BUILDROOT/cloudland-${version}-${release}.${arch}/tmp/cloudland
cp /tmp/cloudland.tar.gz rpmbuild/BUILDROOT/cloudland-${version}-${release}.${arch}/tmp/cloudland
cp /tmp/grpc.tar.gz rpmbuild/BUILDROOT/cloudland-${version}-${release}.${arch}/tmp/cloudland

cat > rpmbuild/SPECS/cloudland.spec <<EOF
Name:           cloudland
Version:        ${version}
Release:        ${release}
Summary:        CloudLand is a light weight IaaS

License:        Apache License 2.0
URL:            https://github.com/IBM/cloudland

%description
Cloudland installer which will copy packed grpc and cloudland to /tmp/cloudland.
There are two packages:
1. grpc.tar.gz: the compiled gRPC libraries which will be unpacked to /usr/local
2. cloudland.tar.gz: the compiled cloudland binaries which include sci, cloudland and libvirt-console-proxy, they will be unpacked to /opt
Commit ID of the CloudLand is $commitID

%files
/tmp/cloudland/grpc.tar.gz
/tmp/cloudland/cloudland.tar.gz

%post
tar xzf /tmp/cloudland/grpc.tar.gz -C /
sudo bash -c 'echo /usr/local/lib > /etc/ld.so.conf.d/protobuf.conf'
sudo ldconfig
tar xzf /tmp/cloudland/cloudland.tar.gz -C /
grep -E "cland" /etc/passwd > /dev/null 2>&1
if [ $? -ne 0 ]; then
    useradd cland
    echo 'cland ALL=(ALL) NOPASSWD:ALL' >> /etc/sudoers.d/cland
fi
chown -R cland:cland /opt/cloudland
chown -R cland:cland /opt/libvirt-console-proxy

%changelog
EOF

rpmbuild -bb rpmbuild/SPECS/cloudland.spec
