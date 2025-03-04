#!/bin/bash

cd `dirname $0`
source ../../cloudrc
[ $# -lt 3 ] && echo "$0 <vm_ID> <rdp_port> <password>" && exit -1

vm_ID=$1
rdp_port=$2
password=$3


TIMEOUT=60
WAIT_TIME=5
ELAPSED_TIME=0

# wait for the VM to boot
log_debug $vm_ID "Waiting for Windows VM '$vm_ID' to boot..."
while true; do
    # check if the VM is running
    VM_STATE=$(virsh domstate "$vm_ID" 2>&1)
    if [ "$VM_STATE" == "running" ]; then
        log_debug $vm_ID "Windows VM '$vm_ID' is running."
        break
    fi

    # check if the timeout has been reached
    if [ $ELAPSED_TIME -ge $TIMEOUT ]; then
        log_debug $vm_ID "Timeout waiting for Windows VM '$vm_ID' to start after $TIMEOUT seconds."
        die "Timeout waiting for Windows VM '$vm_ID' to start after $TIMEOUT seconds."
    fi

    sleep $WAIT_TIME
    ELAPSED_TIME=$((ELAPSED_TIME + WAIT_TIME))
done

# wait for the windows guest agent to start
log_debug $vm_ID "Waiting for Windows VM '$vm_ID' to start the guest agent..."
while true; do
    # check if the guest agent is running
    virsh qemu-agent-command "$vm_ID" '{"execute":"guest-ping"}'
    if [ $? -eq 0 ]; then
        log_debug $vm_ID "Windows VM '$vm_ID' has started the guest agent."
        break
    fi

    # check if the timeout has been reached
    if [ $ELAPSED_TIME -ge $TIMEOUT ]; then
        log_debug $vm_ID "Timeout waiting for Windows VM '$vm_ID' to start the guest agent after $TIMEOUT seconds."
        die "Timeout waiting for Windows VM '$vm_ID' to start the guest agent after $TIMEOUT seconds."
    fi

    sleep $WAIT_TIME
    ELAPSED_TIME=$((ELAPSED_TIME + WAIT_TIME))
done

PS_SCRIPT='Set-ItemProperty -Path \"HKLM:\\SYSTEM\\CurrentControlSet\\Control\\Terminal Server\\WinStations\\RDP-Tcp\" -Name PortNumber -Value '${rdp_port}'; Restart-Service -Name \"TermService\" -Force; New-NetFirewallRule -DisplayName \"RDP-TCP-'${rdp_port}'\" -Action Allow -Protocol TCP -LocalPort '${rdp_port}

ELAPSED_TIME=0
PS_SUCEED_TIMES=0
log_debug $vm_ID "Executing PowerShell script to change RDP port..."
while true; do
    OUTPUT=$(virsh qemu-agent-command "$vm_ID" '{"execute":"guest-exec","arguments":{"path":"C:\\Windows\\System32\\WindowsPowerShell\\v1.0\\powershell.exe","arg":["-Command","'"$PS_SCRIPT"'"],"capture-output":true}}')
    if [ -n "${OUTPUT}" ]; then
        log_debug $vm_ID "$vm_ID exec powershell: $OUTPUT"
        QA_RS=$(jq -r '.return' <<< $OUTPUT)
        if [ -n "${QA_RS}" ]; then
            PS_SUCEED_TIMES=$((PS_SUCEED_TIMES + 1))
            if [ ${PS_SUCEED_TIMES} -gt 2 ]; then
                log_debug $vm_ID "$vm_ID exec powershell succeed"
                break
            fi
        fi
    fi
    if [ $ELAPSED_TIME -ge $TIMEOUT ]; then
        log_debug $vm_ID "$vm_ID timeout while waiting guest agent ready"
        die "Timeout waiting for Windows VM '$vm_ID' to execute PowerShell script after $TIMEOUT seconds."
    fi
    sleep $WAIT_TIME
    ELAPSED_TIME=$((ELAPSED_TIME + WAIT_TIME))
done

# change the password
if [ -n "$password" ]; then
    ELAPSED_TIME=0
    PS_SUCEED_TIMES=0
    log_debug $vm_ID "$vm_ID setting user password"
    while true; do
        OUTPUT=$(virsh set-user-password --domain $vm_ID --user Administrator --password $password 2>&1)
        exit_code=$?
        if [ $exit_code -eq 0 ]; then
            log_debug $vm_ID "$vm_ID set password result: $OUTPUT"
            PS_SUCEED_TIMES=$((PS_SUCEED_TIMES + 1))
            if [ ${PS_SUCEED_TIMES} -gt 5 ]; then
                log_debug $vm_ID "$vm_ID set password succeed"
                break
            fi
        fi
        if [ $ELAPSED_TIME -ge $TIMEOUT ]; then
            log_debug $vm_ID "$vm_ID timeout while set password"
            die "Timeout waiting for Windows VM '$vm_ID' to set password script after $TIMEOUT seconds."
        fi
        sleep $WAIT_TIME
        ELAPSED_TIME=$((ELAPSED_TIME + WAIT_TIME))
    done
fi
