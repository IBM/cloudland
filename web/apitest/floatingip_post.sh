#!/bin/bash

source tokenrc

cat >tmp.json <<EOF
{
}
EOF

curl -k -XPOST -H "Authorization: bearer $token" "$endpoint/api/v1/floating_ips" -d @./tmp.json | jq .
