#!/bin/bash

source tokenrc

cat >tmp.json <<EOF
{
  "power_action": "start"
}
EOF

curl -k -XPATCH -H "Authorization: bearer $token" "$endpoint/api/v1/instances/91f93927-f0d2-4f58-8707-039c1a99a7be" -d @./tmp.json
