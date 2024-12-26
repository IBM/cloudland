#!/bin/bash

source tokenrc

cat >tmp.json <<EOF
{
  "name": "secgroup-new-$RANDOM",
  "is_default": true
}
EOF

curl -k -XPATCH -H "Authorization: bearer $token" -H "X-Resource-User: cathy" -H "X-Resource-Org: cathy" "$endpoint/api/v1/security_groups/628aba55-db62-4802-87ae-cbf382723b2b" -d@./tmp.json | jq .
