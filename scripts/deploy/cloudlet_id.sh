#!/bin/bash

echo SCI_CLIENT_ID=$(hostname -s | sed "s/.*-\(.\)/\1/") >> /etc/default/cloudlet
