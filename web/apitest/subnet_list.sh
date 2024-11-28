#!/bin/bash

source tokenrc

curl -k -XGET -H "Authorization: bearer $token" "$endpoint/api/v1/subnets" | jq .
