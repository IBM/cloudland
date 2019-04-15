#!/bin/bash

for i in {1..180}; do
    rt=$(pidof cloudland)
    [ -z "$rt" ] && break
    sleep 1
done
