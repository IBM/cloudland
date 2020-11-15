#!/bin/bash

# Create new VM
function create_vm()
{
    path=$(cd $(dirname $0) && pwd)
    #echo $path

    source $path/testrc

    cookiefile=$path/cookie.txt
    keyfile=$path/key.txt
    gatewayfile=$path/gateway.txt
    vminfofile=$path/vminfo.txt
    floatingipfile=$path/floatingips.txt
    sshkeyfile=$path/sshkey.txt

    #save cookie.txt
    echo "Admin login!"
    curl -s -c cookie.txt -X POST -d "username=admin&password=passw0rd" --insecure $endpoint/login

    echo "Create key!"
    curl -s -b cookie.txt -X POST -d "name=mykey" --insecure $endpoint/keys/confirm
    curl -s -o key.txt -b cookie.txt -X GET --insecure $endpoint/keys -H "X-Json-Format:yes"
    #mykey=$(cat key.txt | jq -r '.keys[0].ID')
    #echo $mykey

    #echo "Create subnet!"
    #curl -i -g -b cookie.txt -X POST -d "name=ziwang&network=10.4.35.56&netmask=255.255.255.240&dhcp="yes"" --insecure $endpoint/subnets/new

    echo "Create mygateway!"
    curl -s -b cookie.txt -X POST -d "name=mygateway&subnets=3" --insecure $endpoint/gateways/new
    curl -s -o gateway.txt -b cookie.txt -X GET --insecure $endpoint/gateways -H "X-Json-Format:yes"
    mygateway=$(cat gateway.txt | jq '.gateways[0].ID')
    #echo $mygateway

    #echo "Create vm instance!"
    curl -s -o vminfo.txt -b cookie.txt -X POST -d "hostname=centos&hyper=-1&count=1&image=1&flavor=1&primary=3&keys=13" --insecure $endpoint/instances/new -H "X-Json-Format:yes"
    vmstatus=$(cat vminfo.txt | jq -r '.Status')
    vmhostname=$(cat vminfo.txt | jq -r '.Hostname')
    vmid=$(cat vminfo.txt | jq -r '.ID')
    #vmkey=$(cat vminfo.txt | jq '.Keys')
    #echo $vmstatus
    if [ $vmstatus = "pending" ]; then
       echo "Create new VM instance succeeded!"
       echo "New VM's Hostname: $vmhostname"
       echo "New VM's Indexnumber: $vmid"
       #echo "New VM's Key:$vmkey"
    else
       echo "Create new VM instance failed!"
    fi

    echo "Create floatingip!"
    curl -s -b cookie.txt -X POST -d "instance=$vmid&ftype="public"" --insecure $endpoint/floatingips/new
    curl -s -o floatingips.txt -b cookie.txt -X GET --insecure $endpoint/floatingips -H "X-Json-Format:yes"
    floatingip=$(cat floatingips.txt | jq -r '.floatingips[0].FipAddress')
    #echo $floatingip

    fip=${floatingip%/*}
    echo "The floatingip of new VM is: $fip"

    if [ -f "$cookiefile" ]; then
       #echo "Delete cookiefile!"
       rm -f $cookiefile
    fi

    if [ -f "$keyfile" ]; then
       #echo "Delete keyfile!"
       rm -f $keyfile
    fi
    if [ -f "$gatewayfile" ]; then
       #echo "Delete gatewayfile!"
       rm -f $gatewayfile
    fi
    if [ -f "$vminfofile" ]; then
       #echo "Delete vminfofile!"
       rm -f $vminfofile
    fi
    #if [ -f "$floatingipfile" ]; then
        #echo "Delete floatingipfile!"
        #rm -f $floatingipfile
    #fi
}

# Test ssh connect
function test_sshconnect()
{
    path=$(cd $(dirname $0) && pwd)
    #echo $path

    sshkeyfile=$path/sshkey.txt

    floatingip=$(cat $path/floatingips.txt | jq -r '.floatingips[0].FipAddress')
    #echo $floatingip

    fip=${floatingip%/*}
    echo "Test ssh connect by using floatingip: $fip"

    sudo chmod 600 $sshkeyfile

    set timeout 30
    ssh -i $sshkeyfile centos@${fip} true

    if [ $? -ne 0 ]; then
       echo "Connect failed!"
    else
       echo "Connect succeed!"
    fi 
}

create_vm
sleep 20
test_sshconnect

