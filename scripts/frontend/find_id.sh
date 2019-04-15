#!/bin/bash

for i in `pidof scia64` `pidof cloudlet`; do 
    echo $i
    cat /proc/$i/environ | strings | grep SCI_CLIENT_ID
done
