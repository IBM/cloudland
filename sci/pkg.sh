#!/bin/bash
cd $(dirname ${BASH_SOURCE[0]})
dpkg -s build-essential || apt-get install -qq build-essential
./configure --prefix=/opt/sci
make
make install
cp scid/scid@.service /lib/systemd/system/scid@.service
cat > /etc/ld.so.conf.d/extra-x86_64.conf <<EOF
/usr/local/lib
/opt/sci/apple/lib64
/opt/sci/berry/lib64
EOF
mkdir /opt/sci/apple
mv /opt/sci/bin /opt/sci/sbin /opt/sci/lib64 /opt/sci/include /opt/sci/apple
cp -rf /opt/sci/apple /opt/sci/berry
tar cfz sci_apple.tgz /opt/sci/apple \
    /lib/systemd/system/scid@.service \
    /etc/ld.so.conf.d/extra-x86_64.conf
tar cfz sci_berry.tgz /opt/sci/berry \
    /lib/systemd/system/scid@.service \
    /etc/ld.so.conf.d/extra-x86_64.conf
cat > sci_apple.sum <<EOF
[sci_apple]
version = "`git describe --always --tags`"
sha1sum = "`shasum sci_apple.tgz | cut -c 1-40`"
EOF
cat > sci_berry.sum <<EOF
[sci_berry]
version = "`git describe --always --tags`"
sha1sum = "`shasum sci_berry.tgz | cut -c 1-40`"
EOF

