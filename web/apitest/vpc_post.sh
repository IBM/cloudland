#!/bin/bash

source tokenrc

cat >tmp.json <<EOF
{
  "name": "vpc-$RANDOM"
}
EOF

curl -k -XPOST -H "Authorization: bearer $token" "$endpoint/api/v1/vpcs" -d @./tmp.json | jq .
