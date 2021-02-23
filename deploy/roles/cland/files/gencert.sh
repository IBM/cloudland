#!/bin/bash

cert_dir=/opt/libvirt-console-proxy/cert
mkdir -p $cert_dir
cd $cert_dir
if [ -e cacert.pem -a -e cakey.pem -a -e servercert.pem -a -e serverkey.pem ]; then
    exit 0
fi
rm -f /tmp/ca.info
rm -f /tmp/server.info
cat >/tmp/ca.info <<EOF
cn = console-proxy
ca
cert_signing_key
EOF
certtool --generate-privkey --outfile cakey.pem > /dev/null 2>&1
certtool --generate-self-signed --load-privkey cakey.pem --template /tmp/ca.info --outfile cacert.pem > /dev/null 2>&1
certtool --generate-privkey --outfile serverkey.pem > /dev/null 2>&1
cat >/tmp/server.info <<EOF
organization = console-proxy
cn = cloudland
tls_www_server
encryption_key
signing_key
EOF
certtool --generate-certificate --load-privkey serverkey.pem --load-ca-certificate cacert.pem --load-ca-privkey cakey.pem --template /tmp/server.info --outfile servercert.pem > /dev/null 2>&1
rm -f /tmp/ca.info
rm -f /tmp/server.info
