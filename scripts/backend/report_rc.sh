#!/bin/bash

cd `dirname $0`
source ../cloudrc

exec <&-

function init_db 
{
    schema=$(sql_exec ".schema")
    [ -n "$schema" ] && return
    sql_exec "CREATE TABLE cpu(real int, virtual int, dedicated int, avail int, used int, weight int)"
    sql_exec "CREATE TABLE memory(total int, free int, avail int, used int, weight int)"
    sql_exec "CREATE TABLE disk(total int, avail int, used int, weight int)"
    sql_exec "CREATE TABLE network(total int, avail int, used int, weight int)"
    sql_exec "CREATE TABLE router (name varchar(32), internal_ip varchar(32), internal_if varchar(16), external_ip varchar(32), external_if varchar(32), vrrp_vni int, vrrp_ip varchar(32))"
    sql_exec "CREATE TABLE floating (router varchar(32), ext_type varchar(12), interface varchar(16), floating_ip varchar(32), internal_ip varchar(32))"
    sql_exec "CREATE TABLE gateway (router varchar(32), subnet_vni int, interface varchar(16), gateway_ip varchar(32))"
    weight=16
    let real=$(cat /proc/cpuinfo | grep processor | tail -1 | cut -d: -f 2 | xargs)+1
    let virtual=$real*$weight
    sql_exec "INSERT INTO cpu(real, virtual, dedicated, avail, used, weight) values ($real, $virtual, 0, $virtual, 0, $weight)"
    memory=$(free -m | grep Mem:)
    total=$(echo $memory | awk '{print $2}')
    free=$(echo $memory | awk '{print $4}')
    avail=$(echo $memory | awk '{print $7}')
    sql_exec "INSERT INTO memory(total, free, avail, used, weight) values ($total, $free, $avail, 0, 2)"
    disk=$(df $cache_dir | tail -1)
    let total=$(echo $disk | awk '{print $2}')/1024/1024
    let free=$(echo $disk | awk '{print $3}')/1024/1024
    let avail=$(echo $disk | awk '{print $4}')/1024/1024
    sql_exec "INSERT INTO disk(total, avail, used, weight) values ($total, $avail, 0, 2)"
    sql_exec "INSERT INTO network(total, avail, used, weight) values (0, 0, 0, 1)"
}

function get_value()
{
    table=$1
    key=$2
    sql_exec "SELECT $key from $table"
}

function update_cpu()
{
    virtual=$(get_value cpu virtual)
    used=$(get_value cpu used )
    dedicated=$(get_value cpu dedicated)
    let avail=$virtual-$dedicated-$used
    sql_exec "UPDATE cpu set avail='$avail'"
}

function update_memory()
{
    total=$(get_value memory total)
    free=$(get_value memory free)
    used=$(get_value memory used)
#    weight=$(get_value memory weight)
    let avail=($total-$used)
    sql_exec "UPDATE memory set avail='$avail'"
}

function update_disk()
{
    total=$(get_value disk total)
    used=$(get_value disk used)
#    weight=$(get_value disk weight)
    let avail=($total-$used)
    sql_exec "UPDATE disk set avail='$avail'"
}

function update_network()
{
    return
}

function update_resource()
{
    init_db
    update_cpu
    update_memory
    update_disk
    update_network
}

update_resource
cpu=$(sql_exec "SELECT avail from cpu")
total_cpu=$(sql_exec "SELECT virtual from cpu")
memory=$(sql_exec "SELECT avail from memory")
total_memory=$(sql_exec "SELECT total from memory")
disk=$(sql_exec "SELECT avail from disk")
total_disk=$(sql_exec "SELECT total from disk")
network=$(sql_exec "SELECT avail from network")
total_network=$(sql_exec "SELECT total from network")
load=$(w | head -1 | cut -d',' -f5 | cut -d'.' -f1 | xargs)
total_load=0
echo "cpu=$cpu/$total_cpu memory=$memory/$total_memory disk=$disk/$total_disk network=$network/$total_network load=$load/$total_load"
