#!/bin/bash

source tokenrc

cat >tmp.json <<EOF
{
  "hostname": "test",
  "primary_interface": {
    "subnet": {
      "id": "8bc206c8-0ced-49f8-ba9b-4b9717fbacc5"
    },
    "site_subnets": [
        {
            "id": "eb8b4ea9-c885-48c1-ab87-c63804859b71"
        }
    ],
    "inbound": 100,
    "outbound": 100
  },
  "flavor": "small",
  "image": {
    "id": "84d2a640-d6c3-4bff-9e03-5a5a535560c6"
  },
  "keys": [
    {
      "id": "59dd901d-ac7d-4918-afbf-ff485de31f07"
    }
  ],
  "zone": "zone0"
}
EOF

curl -k -XPOST -H "Authorization: bearer $token" "$endpoint/api/v1/instances" -d @./tmp.json | jq .
