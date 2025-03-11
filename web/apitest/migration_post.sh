#!/bin/bash

source tokenrc

cat >tmp.json <<EOF
{
    "name": "migration-$RANDOM",
    "instances": [
        {
            "id": "f0a1db43-25d9-4dac-a02e-3272a0fc75d7"
        }
    ]
}
EOF

curl -k -XPOST -H "Authorization: bearer $token" "$endpoint/api/v1/migrations" -d @./tmp.json | jq .
