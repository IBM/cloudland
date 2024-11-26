#!/bin/bash

source tokenrc

curl -k -XGET -H "Authorization: bearer $token" 'https://127.0.0.1:8255/api/v1/instances' | jq .
curl -k -XGET -H "Authorization: bearer $token" -H "X-Resource-User: cathy" -H "X-Resource-Org: cathy" 'https://127.0.0.1:8255/api/v1/instances' | jq .
