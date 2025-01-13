#!/bin/bash

source tokenrc

cat >tmp.json <<EOF
{
    "name": "fip-$RANDOM",
    "instance": {
        "id": "a40318db-1221-418e-8e01-15f395e96484"
    },
    "inbound": 100,
    "outbound": 100
}
EOF

curl -k -XPOST -H "Authorization: bearer $token" "$endpoint/api/v1/floating_ips" -d @./tmp.json | jq .
