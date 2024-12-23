#!/bin/bash

source tokenrc

cat >tmp.json <<EOF
{
  "hostname": "test",
  "primary_interface": {
    "subnet": {
      "id": "f520c67b-81ad-47b0-b52b-75e593ff1b05"
    }
  },
  "count": 3,
  "flavor": "small",
  "image": {
    "id": "3c0cca59-df4f-4daa-bfa1-04d771e1a17c"
  },
  "keys": [
    {
      "id": "506d75da-1e3f-47a2-8f98-8ff7deefa0f0"
    }
  ],
  "zone": "zone0"
}
EOF

curl -k -XPOST -H "Authorization: bearer $token" "$endpoint/api/v1/instances" -d @./tmp.json | jq .
