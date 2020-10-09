#!/bin/bash

#date
#echo 'hello world'

#delete cookie.txt
path=$(cd $(dirname $0) && pwd)
echo $path

file=$path/cookie.txt
#file="/opt/cloudland/test/cookie.txt"

if [ -f "$file" ]; then
   echo "delete file!"
   rm $file
fi

#save cookie.txt
echo "Save cookie!"
curl -v -c cookie.txt -X POST -d "username=admin&password=passw0rd" http://169.61.25.50/login?redirect_to

#visit by cookie.txt
echo “Login no password!”
curl -i -b cookie.txt http://169.61.25.50/dashboard

#visit webpage
echo “Visit instances webpage!”
curl -i -b cookie.txt  http://169.61.25.50/instances

