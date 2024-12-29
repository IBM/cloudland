#!/bin/bash

source tokenrc

cat >tmp.json <<EOF
{
  "name": "vpc-$RANDOM"
}
EOF
vpc_id=$(curl -k -XPOST -H "Authorization: bearer $token" -H "X-Resource-User: cathy" -H "X-Resource-Org: cathy" "$endpoint/api/v1/vpcs" -d @./tmp.json | jq -r .id)

cat >tmp.json <<EOF
{
  "name": "secgroup-$RANDOM",
  "vpc": {
    "id": "$vpc_id"
  }
}
EOF
secgroup_id=$(curl -k -XPOST -H "Authorization: bearer $token" -H "X-Resource-User: cathy" -H "X-Resource-Org: cathy" "$endpoint/api/v1/security_groups" -d @./tmp.json | jq -r .id)

cat >tmp.json <<EOF
{
  "remote_cidr": "192.168.10.10/32",
  "direction": "ingress",
  "protocol": "tcp",
  "port_min": 443,
  "port_max": 443
}
EOF
curl -k -XPOST -H "Authorization: bearer $token" -H "X-Resource-User: cathy" -H "X-Resource-Org: cathy" "$endpoint/api/v1/security_groups/$secgroup_id/rules" -d @./tmp.json | jq .
