#!/bin/bash

source tokenrc

cat >tmp.json <<EOF
{
  "public_subnet": {
    "id": "36ee20ac-1680-4473-8c66-cf79b31c663f"
  },
  "instance": {
    "id": "4f9c35f2-517e-4d76-97ab-8893b6835e6e"
  }
}
EOF

curl -k -XPOST -H "Authorization: bearer $token" "$endpoint/api/v1/floating_ips" -d @./tmp.json | jq .
