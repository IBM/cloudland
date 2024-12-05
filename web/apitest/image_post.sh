#!/bin/bash

source tokenrc

cat >tmp.json <<EOF
{
  "name": "ubuntu2204-$RANDOM",
  "os_version": "2204",
  "download_url": "https://cloud-images.ubuntu.com/jammy/20241004/jammy-server-cloudimg-amd64.img",
  "user": "ubuntu"
}
EOF

curl -k -XPOST -H "Authorization: bearer $token" "$endpoint/api/v1/images" -d @./tmp.json | jq .
