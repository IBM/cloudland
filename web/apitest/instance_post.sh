#!/bin/bash

source tokenrc

cat >tmp.json <<EOF
{
  "hostname": "test-$RANDOM",
  "primary_interface": {
    "subnet": {
      "id": "679b372c-614e-4365-ad29-3ac169426ca5"
    }
  },
  "secondary_interfaces": [
    {
      "subnet": {
        "id": "5f4c85db-265b-45c6-b7d7-76e389dbe5b9"
      }
    },
    {
      "subnet": {
        "id": "e5c68a54-1b0c-435b-b689-5c18c8e963dd"
      }
    }
  ],
  "flavor": "x-medium",
  "image": {
    "id": "97948d36-279a-4eb4-96a6-32e2348eba3e"
  },
  "keys": [
    {
      "id": "013d50d9-9f5b-46a0-a49f-0e70c1c17745"
    }
  ],
  "zone": "zone0"
}
EOF

curl -k -XPOST -H "Authorization: bearer $token" "$endpoint/api/v1/instances" -d @./tmp.json | jq .
