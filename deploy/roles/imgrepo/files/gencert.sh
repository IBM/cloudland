#!/bin/bash

cat <<EOF > /tmp/server.info
organization = cloudland
tls_www_server
encryption_key
signing_key
EOF

mkdir -p /etc/ssl/private
certtool --generate-privkey --outfile /etc/ssl/private/nginx-selfsigned.key
certtool --generate-self-signed --load-privkey /etc/ssl/private/nginx-selfsigned.key --template /tmp/server.info --outfile /etc/ssl/certs/nginx-selfsigned.crt
openssl dhparam -out /etc/ssl/certs/dhparam.pem 2048
rm -f /tmp/server.info
