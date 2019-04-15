#!/bin/bash

ethcfg=/etc/network/interfaces.d/eth0.cfg
brcfg=/etc/network/interfaces.d/br4090.cfg

cp $ethcfg $brcfg
sed -i "s/eth0/br4090/g" $brcfg
sed -i "/iface br4090 inet/a bridge_ports eth0" $brcfg
ifdown eth0
sed -i "s/static/manual/" $ethcfg
sed -i "3,$ d" $ethcfg
ifup eth0
ifup br4090
