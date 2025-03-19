#!/bin/bash

source tokenrc

cat >tmp.json <<EOF
{
  "username": "user-$RANDOM",
  "password": "test12345ytr"
}
EOF

curl -k -XPOST -H "Authorization: bearer $token" "$endpoint/api/v1/users" -d @./tmp.json | jq .
