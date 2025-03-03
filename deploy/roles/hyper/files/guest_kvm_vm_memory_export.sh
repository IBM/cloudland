#!/bin/bash
while true; do
  echo '# HELP guest_kvm_vm_memory_usage_bytes Memory usage of KVM VMs in bytes'
  echo '# TYPE guest_kvm_vm_memory_usage_bytes gauge'
  for uuid in $(virsh list --uuid); do
    mem_usage=$(virsh dommemstat "$uuid" | awk '/rss/ {print $2}')
    if [[ ! -z "$mem_usage" ]]; then
      echo "guest_kvm_vm_memory_usage_bytes{uuid=\"$uuid\"} $((mem_usage * 1024))"
    fi
  done > /var/lib/node_exporter/guest_kvm_vm_memory_usage.prom
  sleep 15
done