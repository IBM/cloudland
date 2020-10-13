#!/bin/bash

#date
#echo 'hello world'

#delete cookie.txt
path=$(cd $(dirname $0) && pwd)
echo $path

file=$path/cookie.txt
#file="/opt/cloudland/test/cookie.txt"

#if [ -f "$file" ]; then
#   echo "delete file!"
#   rm $file
#fi

#save cookie.txt
echo "Save cookie!"
curl -v -c cookie.txt -X POST -d "username=admin&password=passw0rd" --insecure http://169.61.25.58/login

#echo "Create key!"
#curl -i -g -b cookie.txt -X POST -d "name=key" --insecure http://169.61.25.58/keys/confirm

#echo "Create ziwang!"
#curl -i -g -b cookie.txt -X POST -d "name=ziwang&network=10.4.35.56&netmask=255.255.255.240&dhcp="yes"" --insecure http://169.61.25.58/subnets/new

echo "Create wangguan!"
curl -i -g -b cookie.txt -X POST -d "name=wangguan&subnets=14" --insecure http://169.61.25.58/gateways/new

#echo "Create vm instance!"
#curl -i -g -b cookie.txt -X POST -d "hostname=inst&hyper=-1&count=1&image=1&flavor=1&primary=14" --insecure http://169.61.25.58/instances/new -H "Accept:application/json"

#echo "Create floatingip!"
#curl -i -g -b cookie.txt -X POST -d "instance=16&ftype="public"" --insecure http://169.61.25.58/floatingips/new

if [ -f "$file" ]; then
   echo "delete file!"
   rm $file
fi
