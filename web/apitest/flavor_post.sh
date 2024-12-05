#!/bin/bash

source tokenrc

cat >tmp.json <<EOF
{
  "name": "flavor-$RANDOM",
  "cpu": 1,
  "memory": 256,
  "disk": 20
}
EOF

curl -k -XPOST -H "Authorization: bearer $token" "$endpoint/api/v1/flavors" -d @./tmp.json | jq .
