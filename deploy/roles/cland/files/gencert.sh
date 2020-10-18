#!/bin/bash

cert_dir=/opt/libvirt-console-proxy/cert
cat >/tmp/ca.info <<EOF
cn = console-proxy
ca
cert_signing_key
EOF
mkdir $cert_dir
cat >/tmp/ca.info <<EOF
cn = console-proxy
ca
cert_signing_key
EOF
cd $cert_dir
certtool --generate-privkey >cakey.pem
certtool --generate-self-signed --load-privkey cakey.pem --template /tmp/ca.info --outfile cacert.pem
certtool --generate-privkey >serverkey.pem
cat <<EOF > /tmp/server.info
organization = console-proxy
cn = cloudland
tls_www_server
encryption_key
signing_key
EOF
certtool --generate-certificate --load-privkey serverkey.pem --load-ca-certificate cacert.pem --load-ca-privkey cakey.pem --template /tmp/server.info --outfile servercert.pem
rm -f /tmp/ca.info
