#!/bin/bash

if [ -e /etc/ssl/private/nginx-selfsigned.key -a -e /etc/ssl/certs/dhparam.pem -a -e /etc/ssl/private/nginx-selfsigned.key ]; then
    exit 0
fi
rm -f /tmp/server.info
cat <<EOF > /tmp/server.info
organization = cloudland
tls_www_server
encryption_key
signing_key
EOF

mkdir -p /etc/ssl/private
certtool --generate-privkey --outfile /etc/ssl/private/nginx-selfsigned.key > /dev/null 2>&1
certtool --generate-self-signed --load-privkey /etc/ssl/private/nginx-selfsigned.key --template /tmp/server.info --outfile /etc/ssl/certs/nginx-selfsigned.crt > /dev/null 2>&1
openssl dhparam -out /etc/ssl/certs/dhparam.pem 2048 > /dev/null 2>&1
rm -f /tmp/server.info
