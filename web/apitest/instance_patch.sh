#!/bin/bash

source tokenrc

cat >tmp.json <<EOF
{
  "hostname": "test123",
  "power_action": "start"
}
EOF

curl -k -XPATCH -H "Authorization: bearer $token" "$endpoint/api/v1/instances/867cd561-3347-4f88-b32c-cf729c24f171" -d @./tmp.json
