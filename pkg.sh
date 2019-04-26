#!/bin/bash

cd "$(dirname ${BASH_SOURCE[0]})"

for pkg in git build-essential autoconf libtool pkg-config libtool-bin wget; do
    dpkg -s $pkg &>/dev/null || apt-get install -qq $pkg
done

# install grpc
wget http://169.61.25.53:8080/share/grpc-xenial.tgz
tar xCfz / grpc-xenial.tgz

mkdir -p /opt/cloudland/scripts
# install cloudland
(cd src; ./setup-sci.sh; make;  make install)

# copy scripts
cp -rf scripts/xml scripts/backend scripts/deploy scripts/cloudrc /opt/cloudland/scripts

find /opt/cloudland -type f -name '*.sh' -exec chmod a+x {} \;
/opt/cloudland/scripts/deploy/dirs.sh

mkdir -p build/opt
mkdir -p build/lib/systemd/system

# copy services
cp scripts/deploy/*.service build/lib/systemd/system

# copy grpc
tar xCfz build/ grpc-xenial.tgz

rm -f grpc-xenial.tgz

# build ipcalc
mkdir -p build/usr/local/bin
(rm -rf ipcalc; git clone https://github.com/nmav/ipcalc.git; cd ipcalc; USE_GEOIP=no make)
cp -f ipcalc/ipcalc build/usr/local/bin
chmod a+x build/usr/local/bin/ipcalc
rm -rf ipcalc

# make apple
mkdir /opt/cloudland/apple
mv /opt/cloudland/bin /opt/cloudland/lib64 /opt/cloudland/scripts /opt/cloudland/apple
ln -sf /opt/cloudland/apple/scripts /opt/cloudland/scripts

# copy cloudland
cp -rfP /opt/cloudland build/opt/

# package
tar cCfz build/ cloudland_apple.tgz opt usr lib

# make berry
rm -rf build/opt/cloudland
cp -rf /opt/cloudland/apple /opt/cloudland/berry
ln -sf /opt/cloudland/berry/scripts /opt/cloudland/scripts
cp -rfP /opt/cloudland build/opt/
tar cCfz build/ cloudland_berry.tgz opt usr lib

for name in apple berry; do
    cat > cloudland_${name}.sum <<EOF
[cloudland_${name}]
version = "`git describe --always --tags`"
sha1sum = "`shasum cloudland_${name}.tgz | cut -c 1-40`"
EOF
done

