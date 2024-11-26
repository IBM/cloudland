#!/bin/bash

source tokenrc

curl -k -XGET -H "Authorization: bearer $token" 'https://127.0.0.1:8255/api/v1/floating_ips' | jq .
