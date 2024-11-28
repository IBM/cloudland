#!/bin/bash

source tokenrc

cat >tmp.json <<EOF
{
  "public_ip": "10.0.100.100"
}
EOF

curl -k -XPOST -H "Authorization: bearer $token" "$endpoint/api/v1/floating_ips" -d @./tmp.json | jq .
