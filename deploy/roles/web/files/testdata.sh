#!/bin/bash

passwd=$1
db_host=$2
pub_vlan=$3
pri_vlan=$4
datetime=$(date +'%Y-%m-%d %H:%M:%S')

sqlfile=/tmp/testdata-$(date +%H%M%S).sql
cat >$sqlfile <<EOF
INSERT  INTO "flavors" ("id","created_at","updated_at","deleted_at","name","cpu","memory","disk") VALUES (1, '$datetime','$datetime',NULL,'m1.tiny',1,256,8) RETURNING "flavors"."id";
INSERT  INTO "flavors" ("id","created_at","updated_at","deleted_at","name","cpu","memory","disk") VALUES (2, '$datetime','$datetime',NULL,'m1.normal',4,6144,120) RETURNING "flavors"."id";
ALTER sequence "flavors_id_seq" restart with 3;
INSERT  INTO "images" ("id","created_at","updated_at","deleted_at","name","os_code","format","architecture","status","href","checksum") VALUES (1,'$datetime','$datetime',NULL,'centos7','core','qcow2','x86_64','available','','') RETURNING "images"."id";
INSERT  INTO "images" ("id","created_at","updated_at","deleted_at","name","os_code","format","architecture","status","href","checksum") VALUES (2,'$datetime','$datetime',NULL,'ubuntu1804','bionic','qcow2','x86_64','available','','') RETURNING "images"."id";
INSERT  INTO "images" ("id","created_at","updated_at","deleted_at","name","os_code","format","architecture","status","href","checksum") VALUES (3,'$datetime','$datetime',NULL,'ubuntu1604','xenial','qcow2','x86_64','available','','') RETURNING "images"."id";
ALTER sequence "images_id_seq" restart with 4;
INSERT  INTO "subnets" ("id","created_at","updated_at","deleted_at","name","network","netmask","gateway","start","end","vlan","type","router") VALUES (1,'$datetime','$datetime',NULL,'public','192.168.71.0','255.255.255.0','192.168.71.1/24','192.168.71.100','192.168.71.150',$pub_vlan,'public',0) RETURNING "subnets"."id";
INSERT  INTO "addresses" ("created_at","updated_at","deleted_at","address","netmask","type","subnet_id","interface") VALUES ('$datetime','$datetime',NULL,'192.168.71.100/24','255.255.255.0','ipv4',1,0);
INSERT  INTO "addresses" ("created_at","updated_at","deleted_at","address","netmask","type","subnet_id","interface") VALUES ('$datetime','$datetime',NULL,'192.168.71.101/24','255.255.255.0','ipv4',1,0);
INSERT  INTO "addresses" ("created_at","updated_at","deleted_at","address","netmask","type","subnet_id","interface") VALUES ('$datetime','$datetime',NULL,'192.168.71.102/24','255.255.255.0','ipv4',1,0);
INSERT  INTO "addresses" ("created_at","updated_at","deleted_at","address","netmask","type","subnet_id","interface") VALUES ('$datetime','$datetime',NULL,'192.168.71.103/24','255.255.255.0','ipv4',1,0);
INSERT  INTO "addresses" ("created_at","updated_at","deleted_at","address","netmask","type","subnet_id","interface") VALUES ('$datetime','$datetime',NULL,'192.168.71.104/24','255.255.255.0','ipv4',1,0);
INSERT  INTO "addresses" ("created_at","updated_at","deleted_at","address","netmask","type","subnet_id","interface") VALUES ('$datetime','$datetime',NULL,'192.168.71.105/24','255.255.255.0','ipv4',1,0);
INSERT  INTO "addresses" ("created_at","updated_at","deleted_at","address","netmask","type","subnet_id","interface") VALUES ('$datetime','$datetime',NULL,'192.168.71.106/24','255.255.255.0','ipv4',1,0);
INSERT  INTO "addresses" ("created_at","updated_at","deleted_at","address","netmask","type","subnet_id","interface") VALUES ('$datetime','$datetime',NULL,'192.168.71.107/24','255.255.255.0','ipv4',1,0);
INSERT  INTO "addresses" ("created_at","updated_at","deleted_at","address","netmask","type","subnet_id","interface") VALUES ('$datetime','$datetime',NULL,'192.168.71.108/24','255.255.255.0','ipv4',1,0);
INSERT  INTO "addresses" ("created_at","updated_at","deleted_at","address","netmask","type","subnet_id","interface") VALUES ('$datetime','$datetime',NULL,'192.168.71.109/24','255.255.255.0','ipv4',1,0);
INSERT  INTO "addresses" ("created_at","updated_at","deleted_at","address","netmask","type","subnet_id","interface") VALUES ('$datetime','$datetime',NULL,'192.168.71.110/24','255.255.255.0','ipv4',1,0);
INSERT  INTO "addresses" ("created_at","updated_at","deleted_at","address","netmask","type","subnet_id","interface") VALUES ('$datetime','$datetime',NULL,'192.168.71.111/24','255.255.255.0','ipv4',1,0);
INSERT  INTO "addresses" ("created_at","updated_at","deleted_at","address","netmask","type","subnet_id","interface") VALUES ('$datetime','$datetime',NULL,'192.168.71.112/24','255.255.255.0','ipv4',1,0);
INSERT  INTO "addresses" ("created_at","updated_at","deleted_at","address","netmask","type","subnet_id","interface") VALUES ('$datetime','$datetime',NULL,'192.168.71.113/24','255.255.255.0','ipv4',1,0);
INSERT  INTO "addresses" ("created_at","updated_at","deleted_at","address","netmask","type","subnet_id","interface") VALUES ('$datetime','$datetime',NULL,'192.168.71.114/24','255.255.255.0','ipv4',1,0);
INSERT  INTO "addresses" ("created_at","updated_at","deleted_at","address","netmask","type","subnet_id","interface") VALUES ('$datetime','$datetime',NULL,'192.168.71.115/24','255.255.255.0','ipv4',1,0);
INSERT  INTO "addresses" ("created_at","updated_at","deleted_at","address","netmask","type","subnet_id","interface") VALUES ('$datetime','$datetime',NULL,'192.168.71.116/24','255.255.255.0','ipv4',1,0);
INSERT  INTO "addresses" ("created_at","updated_at","deleted_at","address","netmask","type","subnet_id","interface") VALUES ('$datetime','$datetime',NULL,'192.168.71.117/24','255.255.255.0','ipv4',1,0);
INSERT  INTO "addresses" ("created_at","updated_at","deleted_at","address","netmask","type","subnet_id","interface") VALUES ('$datetime','$datetime',NULL,'192.168.71.118/24','255.255.255.0','ipv4',1,0);
INSERT  INTO "addresses" ("created_at","updated_at","deleted_at","address","netmask","type","subnet_id","interface") VALUES ('$datetime','$datetime',NULL,'192.168.71.119/24','255.255.255.0','ipv4',1,0);
INSERT  INTO "addresses" ("created_at","updated_at","deleted_at","address","netmask","type","subnet_id","interface") VALUES ('$datetime','$datetime',NULL,'192.168.71.120/24','255.255.255.0','ipv4',1,0);
INSERT  INTO "addresses" ("created_at","updated_at","deleted_at","address","netmask","type","subnet_id","interface") VALUES ('$datetime','$datetime',NULL,'192.168.71.121/24','255.255.255.0','ipv4',1,0);
INSERT  INTO "addresses" ("created_at","updated_at","deleted_at","address","netmask","type","subnet_id","interface") VALUES ('$datetime','$datetime',NULL,'192.168.71.122/24','255.255.255.0','ipv4',1,0);
INSERT  INTO "addresses" ("created_at","updated_at","deleted_at","address","netmask","type","subnet_id","interface") VALUES ('$datetime','$datetime',NULL,'192.168.71.123/24','255.255.255.0','ipv4',1,0);
INSERT  INTO "addresses" ("created_at","updated_at","deleted_at","address","netmask","type","subnet_id","interface") VALUES ('$datetime','$datetime',NULL,'192.168.71.124/24','255.255.255.0','ipv4',1,0);
INSERT  INTO "addresses" ("created_at","updated_at","deleted_at","address","netmask","type","subnet_id","interface") VALUES ('$datetime','$datetime',NULL,'192.168.71.125/24','255.255.255.0','ipv4',1,0);
INSERT  INTO "addresses" ("created_at","updated_at","deleted_at","address","netmask","type","subnet_id","interface") VALUES ('$datetime','$datetime',NULL,'192.168.71.126/24','255.255.255.0','ipv4',1,0);
INSERT  INTO "addresses" ("created_at","updated_at","deleted_at","address","netmask","type","subnet_id","interface") VALUES ('$datetime','$datetime',NULL,'192.168.71.127/24','255.255.255.0','ipv4',1,0);
INSERT  INTO "addresses" ("created_at","updated_at","deleted_at","address","netmask","type","subnet_id","interface") VALUES ('$datetime','$datetime',NULL,'192.168.71.128/24','255.255.255.0','ipv4',1,0);
INSERT  INTO "addresses" ("created_at","updated_at","deleted_at","address","netmask","type","subnet_id","interface") VALUES ('$datetime','$datetime',NULL,'192.168.71.129/24','255.255.255.0','ipv4',1,0);
INSERT  INTO "addresses" ("created_at","updated_at","deleted_at","address","netmask","type","subnet_id","interface") VALUES ('$datetime','$datetime',NULL,'192.168.71.130/24','255.255.255.0','ipv4',1,0);
INSERT  INTO "addresses" ("created_at","updated_at","deleted_at","address","netmask","type","subnet_id","interface") VALUES ('$datetime','$datetime',NULL,'192.168.71.131/24','255.255.255.0','ipv4',1,0);
INSERT  INTO "addresses" ("created_at","updated_at","deleted_at","address","netmask","type","subnet_id","interface") VALUES ('$datetime','$datetime',NULL,'192.168.71.132/24','255.255.255.0','ipv4',1,0);
INSERT  INTO "addresses" ("created_at","updated_at","deleted_at","address","netmask","type","subnet_id","interface") VALUES ('$datetime','$datetime',NULL,'192.168.71.133/24','255.255.255.0','ipv4',1,0);
INSERT  INTO "addresses" ("created_at","updated_at","deleted_at","address","netmask","type","subnet_id","interface") VALUES ('$datetime','$datetime',NULL,'192.168.71.134/24','255.255.255.0','ipv4',1,0);
INSERT  INTO "addresses" ("created_at","updated_at","deleted_at","address","netmask","type","subnet_id","interface") VALUES ('$datetime','$datetime',NULL,'192.168.71.135/24','255.255.255.0','ipv4',1,0);
INSERT  INTO "addresses" ("created_at","updated_at","deleted_at","address","netmask","type","subnet_id","interface") VALUES ('$datetime','$datetime',NULL,'192.168.71.136/24','255.255.255.0','ipv4',1,0);
INSERT  INTO "addresses" ("created_at","updated_at","deleted_at","address","netmask","type","subnet_id","interface") VALUES ('$datetime','$datetime',NULL,'192.168.71.137/24','255.255.255.0','ipv4',1,0);
INSERT  INTO "addresses" ("created_at","updated_at","deleted_at","address","netmask","type","subnet_id","interface") VALUES ('$datetime','$datetime',NULL,'192.168.71.138/24','255.255.255.0','ipv4',1,0);
INSERT  INTO "addresses" ("created_at","updated_at","deleted_at","address","netmask","type","subnet_id","interface") VALUES ('$datetime','$datetime',NULL,'192.168.71.139/24','255.255.255.0','ipv4',1,0);
INSERT  INTO "addresses" ("created_at","updated_at","deleted_at","address","netmask","type","subnet_id","interface") VALUES ('$datetime','$datetime',NULL,'192.168.71.140/24','255.255.255.0','ipv4',1,0);
INSERT  INTO "addresses" ("created_at","updated_at","deleted_at","address","netmask","type","subnet_id","interface") VALUES ('$datetime','$datetime',NULL,'192.168.71.141/24','255.255.255.0','ipv4',1,0);
INSERT  INTO "addresses" ("created_at","updated_at","deleted_at","address","netmask","type","subnet_id","interface") VALUES ('$datetime','$datetime',NULL,'192.168.71.142/24','255.255.255.0','ipv4',1,0);
INSERT  INTO "addresses" ("created_at","updated_at","deleted_at","address","netmask","type","subnet_id","interface") VALUES ('$datetime','$datetime',NULL,'192.168.71.143/24','255.255.255.0','ipv4',1,0);
INSERT  INTO "addresses" ("created_at","updated_at","deleted_at","address","netmask","type","subnet_id","interface") VALUES ('$datetime','$datetime',NULL,'192.168.71.144/24','255.255.255.0','ipv4',1,0);
INSERT  INTO "addresses" ("created_at","updated_at","deleted_at","address","netmask","type","subnet_id","interface") VALUES ('$datetime','$datetime',NULL,'192.168.71.145/24','255.255.255.0','ipv4',1,0);
INSERT  INTO "addresses" ("created_at","updated_at","deleted_at","address","netmask","type","subnet_id","interface") VALUES ('$datetime','$datetime',NULL,'192.168.71.146/24','255.255.255.0','ipv4',1,0);
INSERT  INTO "addresses" ("created_at","updated_at","deleted_at","address","netmask","type","subnet_id","interface") VALUES ('$datetime','$datetime',NULL,'192.168.71.147/24','255.255.255.0','ipv4',1,0);
INSERT  INTO "addresses" ("created_at","updated_at","deleted_at","address","netmask","type","subnet_id","interface") VALUES ('$datetime','$datetime',NULL,'192.168.71.148/24','255.255.255.0','ipv4',1,0);
INSERT  INTO "addresses" ("created_at","updated_at","deleted_at","address","netmask","type","subnet_id","interface") VALUES ('$datetime','$datetime',NULL,'192.168.71.149/24','255.255.255.0','ipv4',1,0);
INSERT  INTO "addresses" ("created_at","updated_at","deleted_at","address","netmask","type","subnet_id","interface") VALUES ('$datetime','$datetime',NULL,'192.168.71.150/24','255.255.255.0','ipv4',1,0);
INSERT  INTO "subnets" ("id","created_at","updated_at","deleted_at","name","network","netmask","gateway","start","end","vlan","type","router") VALUES (2,'$datetime','$datetime',NULL,'private','172.16.20.0','255.255.255.0','172.16.20.1/24','172.16.20.100','172.16.20.150',$pri_vlan,'private',0) RETURNING "subnets"."id";
INSERT  INTO "addresses" ("created_at","updated_at","deleted_at","address","netmask","type","subnet_id","interface") VALUES ('$datetime','$datetime',NULL,'172.16.20.100/24','255.255.255.0','ipv4',2,0);
INSERT  INTO "addresses" ("created_at","updated_at","deleted_at","address","netmask","type","subnet_id","interface") VALUES ('$datetime','$datetime',NULL,'172.16.20.101/24','255.255.255.0','ipv4',2,0);
INSERT  INTO "addresses" ("created_at","updated_at","deleted_at","address","netmask","type","subnet_id","interface") VALUES ('$datetime','$datetime',NULL,'172.16.20.102/24','255.255.255.0','ipv4',2,0);
INSERT  INTO "addresses" ("created_at","updated_at","deleted_at","address","netmask","type","subnet_id","interface") VALUES ('$datetime','$datetime',NULL,'172.16.20.103/24','255.255.255.0','ipv4',2,0);
INSERT  INTO "addresses" ("created_at","updated_at","deleted_at","address","netmask","type","subnet_id","interface") VALUES ('$datetime','$datetime',NULL,'172.16.20.104/24','255.255.255.0','ipv4',2,0);
INSERT  INTO "addresses" ("created_at","updated_at","deleted_at","address","netmask","type","subnet_id","interface") VALUES ('$datetime','$datetime',NULL,'172.16.20.105/24','255.255.255.0','ipv4',2,0);
INSERT  INTO "addresses" ("created_at","updated_at","deleted_at","address","netmask","type","subnet_id","interface") VALUES ('$datetime','$datetime',NULL,'172.16.20.106/24','255.255.255.0','ipv4',2,0);
INSERT  INTO "addresses" ("created_at","updated_at","deleted_at","address","netmask","type","subnet_id","interface") VALUES ('$datetime','$datetime',NULL,'172.16.20.107/24','255.255.255.0','ipv4',2,0);
INSERT  INTO "addresses" ("created_at","updated_at","deleted_at","address","netmask","type","subnet_id","interface") VALUES ('$datetime','$datetime',NULL,'172.16.20.108/24','255.255.255.0','ipv4',2,0);
INSERT  INTO "addresses" ("created_at","updated_at","deleted_at","address","netmask","type","subnet_id","interface") VALUES ('$datetime','$datetime',NULL,'172.16.20.109/24','255.255.255.0','ipv4',2,0);
INSERT  INTO "addresses" ("created_at","updated_at","deleted_at","address","netmask","type","subnet_id","interface") VALUES ('$datetime','$datetime',NULL,'172.16.20.110/24','255.255.255.0','ipv4',2,0);
INSERT  INTO "addresses" ("created_at","updated_at","deleted_at","address","netmask","type","subnet_id","interface") VALUES ('$datetime','$datetime',NULL,'172.16.20.111/24','255.255.255.0','ipv4',2,0);
INSERT  INTO "addresses" ("created_at","updated_at","deleted_at","address","netmask","type","subnet_id","interface") VALUES ('$datetime','$datetime',NULL,'172.16.20.112/24','255.255.255.0','ipv4',2,0);
INSERT  INTO "addresses" ("created_at","updated_at","deleted_at","address","netmask","type","subnet_id","interface") VALUES ('$datetime','$datetime',NULL,'172.16.20.113/24','255.255.255.0','ipv4',2,0);
INSERT  INTO "addresses" ("created_at","updated_at","deleted_at","address","netmask","type","subnet_id","interface") VALUES ('$datetime','$datetime',NULL,'172.16.20.114/24','255.255.255.0','ipv4',2,0);
INSERT  INTO "addresses" ("created_at","updated_at","deleted_at","address","netmask","type","subnet_id","interface") VALUES ('$datetime','$datetime',NULL,'172.16.20.115/24','255.255.255.0','ipv4',2,0);
INSERT  INTO "addresses" ("created_at","updated_at","deleted_at","address","netmask","type","subnet_id","interface") VALUES ('$datetime','$datetime',NULL,'172.16.20.116/24','255.255.255.0','ipv4',2,0);
INSERT  INTO "addresses" ("created_at","updated_at","deleted_at","address","netmask","type","subnet_id","interface") VALUES ('$datetime','$datetime',NULL,'172.16.20.117/24','255.255.255.0','ipv4',2,0);
INSERT  INTO "addresses" ("created_at","updated_at","deleted_at","address","netmask","type","subnet_id","interface") VALUES ('$datetime','$datetime',NULL,'172.16.20.118/24','255.255.255.0','ipv4',2,0);
INSERT  INTO "addresses" ("created_at","updated_at","deleted_at","address","netmask","type","subnet_id","interface") VALUES ('$datetime','$datetime',NULL,'172.16.20.119/24','255.255.255.0','ipv4',2,0);
INSERT  INTO "addresses" ("created_at","updated_at","deleted_at","address","netmask","type","subnet_id","interface") VALUES ('$datetime','$datetime',NULL,'172.16.20.120/24','255.255.255.0','ipv4',2,0);
INSERT  INTO "addresses" ("created_at","updated_at","deleted_at","address","netmask","type","subnet_id","interface") VALUES ('$datetime','$datetime',NULL,'172.16.20.121/24','255.255.255.0','ipv4',2,0);
INSERT  INTO "addresses" ("created_at","updated_at","deleted_at","address","netmask","type","subnet_id","interface") VALUES ('$datetime','$datetime',NULL,'172.16.20.122/24','255.255.255.0','ipv4',2,0);
INSERT  INTO "addresses" ("created_at","updated_at","deleted_at","address","netmask","type","subnet_id","interface") VALUES ('$datetime','$datetime',NULL,'172.16.20.123/24','255.255.255.0','ipv4',2,0);
INSERT  INTO "addresses" ("created_at","updated_at","deleted_at","address","netmask","type","subnet_id","interface") VALUES ('$datetime','$datetime',NULL,'172.16.20.124/24','255.255.255.0','ipv4',2,0);
INSERT  INTO "addresses" ("created_at","updated_at","deleted_at","address","netmask","type","subnet_id","interface") VALUES ('$datetime','$datetime',NULL,'172.16.20.125/24','255.255.255.0','ipv4',2,0);
INSERT  INTO "addresses" ("created_at","updated_at","deleted_at","address","netmask","type","subnet_id","interface") VALUES ('$datetime','$datetime',NULL,'172.16.20.126/24','255.255.255.0','ipv4',2,0);
INSERT  INTO "addresses" ("created_at","updated_at","deleted_at","address","netmask","type","subnet_id","interface") VALUES ('$datetime','$datetime',NULL,'172.16.20.127/24','255.255.255.0','ipv4',2,0);
INSERT  INTO "addresses" ("created_at","updated_at","deleted_at","address","netmask","type","subnet_id","interface") VALUES ('$datetime','$datetime',NULL,'172.16.20.128/24','255.255.255.0','ipv4',2,0);
INSERT  INTO "addresses" ("created_at","updated_at","deleted_at","address","netmask","type","subnet_id","interface") VALUES ('$datetime','$datetime',NULL,'172.16.20.129/24','255.255.255.0','ipv4',2,0);
INSERT  INTO "addresses" ("created_at","updated_at","deleted_at","address","netmask","type","subnet_id","interface") VALUES ('$datetime','$datetime',NULL,'172.16.20.130/24','255.255.255.0','ipv4',2,0);
INSERT  INTO "addresses" ("created_at","updated_at","deleted_at","address","netmask","type","subnet_id","interface") VALUES ('$datetime','$datetime',NULL,'172.16.20.131/24','255.255.255.0','ipv4',2,0);
INSERT  INTO "addresses" ("created_at","updated_at","deleted_at","address","netmask","type","subnet_id","interface") VALUES ('$datetime','$datetime',NULL,'172.16.20.132/24','255.255.255.0','ipv4',2,0);
INSERT  INTO "addresses" ("created_at","updated_at","deleted_at","address","netmask","type","subnet_id","interface") VALUES ('$datetime','$datetime',NULL,'172.16.20.133/24','255.255.255.0','ipv4',2,0);
INSERT  INTO "addresses" ("created_at","updated_at","deleted_at","address","netmask","type","subnet_id","interface") VALUES ('$datetime','$datetime',NULL,'172.16.20.134/24','255.255.255.0','ipv4',2,0);
INSERT  INTO "addresses" ("created_at","updated_at","deleted_at","address","netmask","type","subnet_id","interface") VALUES ('$datetime','$datetime',NULL,'172.16.20.135/24','255.255.255.0','ipv4',2,0);
INSERT  INTO "addresses" ("created_at","updated_at","deleted_at","address","netmask","type","subnet_id","interface") VALUES ('$datetime','$datetime',NULL,'172.16.20.136/24','255.255.255.0','ipv4',2,0);
INSERT  INTO "addresses" ("created_at","updated_at","deleted_at","address","netmask","type","subnet_id","interface") VALUES ('$datetime','$datetime',NULL,'172.16.20.137/24','255.255.255.0','ipv4',2,0);
INSERT  INTO "addresses" ("created_at","updated_at","deleted_at","address","netmask","type","subnet_id","interface") VALUES ('$datetime','$datetime',NULL,'172.16.20.138/24','255.255.255.0','ipv4',2,0);
INSERT  INTO "addresses" ("created_at","updated_at","deleted_at","address","netmask","type","subnet_id","interface") VALUES ('$datetime','$datetime',NULL,'172.16.20.139/24','255.255.255.0','ipv4',2,0);
INSERT  INTO "addresses" ("created_at","updated_at","deleted_at","address","netmask","type","subnet_id","interface") VALUES ('$datetime','$datetime',NULL,'172.16.20.140/24','255.255.255.0','ipv4',2,0);
INSERT  INTO "addresses" ("created_at","updated_at","deleted_at","address","netmask","type","subnet_id","interface") VALUES ('$datetime','$datetime',NULL,'172.16.20.141/24','255.255.255.0','ipv4',2,0);
INSERT  INTO "addresses" ("created_at","updated_at","deleted_at","address","netmask","type","subnet_id","interface") VALUES ('$datetime','$datetime',NULL,'172.16.20.142/24','255.255.255.0','ipv4',2,0);
INSERT  INTO "addresses" ("created_at","updated_at","deleted_at","address","netmask","type","subnet_id","interface") VALUES ('$datetime','$datetime',NULL,'172.16.20.143/24','255.255.255.0','ipv4',2,0);
INSERT  INTO "addresses" ("created_at","updated_at","deleted_at","address","netmask","type","subnet_id","interface") VALUES ('$datetime','$datetime',NULL,'172.16.20.144/24','255.255.255.0','ipv4',2,0);
INSERT  INTO "addresses" ("created_at","updated_at","deleted_at","address","netmask","type","subnet_id","interface") VALUES ('$datetime','$datetime',NULL,'172.16.20.145/24','255.255.255.0','ipv4',2,0);
INSERT  INTO "addresses" ("created_at","updated_at","deleted_at","address","netmask","type","subnet_id","interface") VALUES ('$datetime','$datetime',NULL,'172.16.20.146/24','255.255.255.0','ipv4',2,0);
INSERT  INTO "addresses" ("created_at","updated_at","deleted_at","address","netmask","type","subnet_id","interface") VALUES ('$datetime','$datetime',NULL,'172.16.20.147/24','255.255.255.0','ipv4',2,0);
INSERT  INTO "addresses" ("created_at","updated_at","deleted_at","address","netmask","type","subnet_id","interface") VALUES ('$datetime','$datetime',NULL,'172.16.20.148/24','255.255.255.0','ipv4',2,0);
INSERT  INTO "addresses" ("created_at","updated_at","deleted_at","address","netmask","type","subnet_id","interface") VALUES ('$datetime','$datetime',NULL,'172.16.20.149/24','255.255.255.0','ipv4',2,0);
INSERT  INTO "addresses" ("created_at","updated_at","deleted_at","address","netmask","type","subnet_id","interface") VALUES ('$datetime','$datetime',NULL,'172.16.20.150/24','255.255.255.0','ipv4',2,0);
ALTER sequence "subnets_id_seq" restart with 3;
EOF

sleep 6
export PGUSER=postgres
export PGPASSWORD=$passwd
export PGHOST=$db_host
export PGDATABASE=hypercube
psql -c "select count(*) from flavors"

psql -v ON_ERROR_STOP=1 -f $sqlfile
rm -f $sqlfile
