#!/bin/bash

source tokenrc

cat >tmp.json <<EOF
{
  "name": "key-$RANDOM",
  "public_key": "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQDPNGok93F8A1sv6rt2EOto4fyB7Ve5RGOWzZ3ztgS4eVQKfOy7Elbb34D7JZnxcWWvwNNSEjG+Lp1ckKOzoHuq3/3AzWXOcl75bH8UJ3K/VG8+XFJbVK5FotYcEXaIcRbwtrCAvZY/1lD41ooW+VOYgLnYDp9dM0ndYCyNofTZK17Odck46GBuA5i3sjFckdQdsR/4/5kRwGQSJuYAoBj4VlXyqsCUsLKH95UF1Jcd8xNROGL8gneKHCCAqmNesdRjrqK0nUhWc6jtqOYFQilWy736LEanDvGMqTlky9xt2x4W19olGiFKFhkl3NVC/A4IbeR14hd4LxqvvCDP581T"
}
EOF

curl -k -XPOST -H "Authorization: bearer $token" "$endpoint/api/v1/keys" -d @./tmp.json | jq .
