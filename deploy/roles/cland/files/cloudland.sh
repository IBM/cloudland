#!/bin/bash

[ -z "$MANAGEMENT_VIP" ] && MANAGEMENT_VIP=127.0.0.1
while true; do
    ip -o addr | grep "\<$MANAGEMENT_VIP\>"
    if [ $? -eq 0 ]; then
        pid=$(pidof cloudland)
        [ -z "$pid" ] && /opt/cloudland/bin/cloudland &
    else
        pid=$(pidof cloudland)
        [ -n "$pid" ] && kill $pid
    fi
    sleep 5
done
