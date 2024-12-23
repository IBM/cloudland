#!/bin/bash

source tokenrc

cat >tmp.json <<EOF
{
  "instance": {
    "id": "624132f5-8d20-48bc-917d-f59952b9ff31"
  }
}
EOF

curl -k -XPATCH -H "Authorization: bearer $token" "$endpoint/api/v1/floating_ips/a40f58e3-5a59-411e-9b39-ddf1b2a68e50" -d @./tmp.json | jq .
