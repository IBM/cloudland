#!/bin/bash

cat <<EOF > /tmp/server.info
organization = cloudland
tls_www_server
encryption_key
signing_key
EOF

certtool --generate-privkey --outfile /etc/ssl/private/cland-selfsigned.key
certtool --generate-self-signed --load-privkey /etc/ssl/private/cland-selfsigned.key --template /tmp/server.info --outfile /etc/ssl/certs/cland-selfsigned.crt
rm -f /tmp/server.info
