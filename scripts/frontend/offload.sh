for iface in eth4 eth6; do
ethtool --offload $iface tso on
ethtool --offload $iface gso off
ethtool --offload $iface gro off
ethtool --offload $iface lro off   
ethtool --offload $iface rxvlan off
ethtool --offload $iface txvlan off
ethtool --offload $iface rxhash off
ethtool --offload $iface l2-fwd-offload on
ethtool --offload $iface hw-tc-offload on
done
