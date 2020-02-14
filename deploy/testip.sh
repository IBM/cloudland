#!/bin/bash

for ns in $(/usr/sbin/ip netns | awk '{print $1}'); do
    ip netns exec $ns /usr/sbin/ip addr show | grep 9.123.120.236
    [ $? -eq 0 ] && echo AAAAAAAAAAA $ns
done
