#!/bin/bash

source tokenrc

cat >tmp.json <<EOF
{
    "name": "fip-$RANDOM"
}
EOF

curl -k -XPOST -H "Authorization: bearer $token" "$endpoint/api/v1/floating_ips" -d @./tmp.json | jq .
