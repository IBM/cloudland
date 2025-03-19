#!/bin/bash

source tokenrc

curl -k -XGET -H "Authorization: bearer $token" "$endpoint/api/v1/migrations/e0c27a5f-d896-4c8f-ac8a-1e09f24c8de3" | jq .
