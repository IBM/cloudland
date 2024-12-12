#!/bin/bash

source tokenrc

cat >tmp.json <<EOF
{
  "public_subnet": {
    "id": "36ee20ac-1680-4473-8c66-cf79b31c663f"
  },
  "instance": {
    "id": "36ee20ac-1680-4473-8c66-cf79b31c663f"
  }
}
EOF

curl -k -XPOST -H "Authorization: bearer $token" "$endpoint/api/v1/floating_ips" -d @./tmp.json | jq .
