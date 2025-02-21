#!/bin/bash

cd `dirname $0`
source ../../cloudrc
[ $# -lt 2 ] && echo "$0 <vm_ID> <rdp_port>" && exit -1

vm_ID=$1
rdp_port=$2


TIMEOUT=60
WAIT_TIME=5
ELAPSED_TIME=0

# wait for the VM to boot
# echo "Waiting for Windows VM '$vm_ID' to boot..."
while true; do
    # check if the VM is running
    VM_STATE=$(virsh domstate "$vm_ID" 2>&1)
    if [ "$VM_STATE" == "running" ]; then
        # echo "Windows VM '$vm_ID' is running."
        break
    fi

    # check if the timeout has been reached
    if [ $ELAPSED_TIME -ge $TIMEOUT ]; then
        # echo "Timeout waiting for Windows VM '$vm_ID' to start after $TIMEOUT seconds."
        exit 1
    fi

    sleep $WAIT_TIME
    ELAPSED_TIME=$((ELAPSED_TIME + WAIT_TIME))
done

# wait for the windows guest agent to start
# echo "Waiting for Windows VM '$vm_ID' to start the guest agent..."
while true; do
    # check if the guest agent is running
    virsh qemu-agent-command "$vm_ID" '{"execute":"guest-ping"}'
    if [ $? -eq 0 ]; then
        # echo "Windows VM '$vm_ID' has started the guest agent."
        break
    fi

    # check if the timeout has been reached
    if [ $ELAPSED_TIME -ge $TIMEOUT ]; then
        exit 1
    fi

    sleep $WAIT_TIME
    ELAPSED_TIME=$((ELAPSED_TIME + WAIT_TIME))
done

PS_SCRIPT='Set-ItemProperty -Path \"HKLM:\\SYSTEM\\CurrentControlSet\\Control\\Terminal Server\\WinStations\\RDP-Tcp\" -Name PortNumber -Value '${rdp_port}'; Restart-Service -Name \"TermService\" -Force; New-NetFirewallRule -DisplayName \"RDP-TCP-'${rdp_port}'\" -Action Allow -Protocol TCP -LocalPort '${rdp_port}

echo "Executing PowerShell script to change RDP port..."
OUTPUT=$(virsh qemu-agent-command "$vm_ID" '{"execute":"guest-exec","arguments":{"path":"C:\\Windows\\System32\\WindowsPowerShell\\v1.0\\powershell.exe","arg":["-Command","'"$PS_SCRIPT"'"],"capture-output":true}}')
