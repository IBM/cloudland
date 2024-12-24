#!/bin/bash

source tokenrc

cat >tmp.json <<EOF
{
  "name": "vpc-$RANDOM"
}
EOF

vpc_id=$(curl -k -XPOST -H "Authorization: bearer $token" -H "X-Resource-User: cathy" -H "X-Resource-Org: cathy" "$endpoint/api/v1/vpcs" -d @./tmp.json | jq -r .id)
echo vpc $vpc_id created

cat >tmp.json <<EOF
{
  "name": "subnet-$RANDOM",
  "network_cidr": "10.240.$(($RANDOM%234)).0/24",
  "vpc": {
    "id": "$vpc_id"
  }
}
EOF

curl -k -XPOST -H "Authorization: bearer $token" -H "X-Resource-User: cathy" -H "X-Resource-Org: cathy" "$endpoint/api/v1/subnets" -d @./tmp.json | jq .
