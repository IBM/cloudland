#!/bin/bash

cd `dirname $0`
source ../../cloudrc
[ $# -lt 3 ] && echo "$0 <vm_ID> <ssh_port> <password>" && exit -1

vm_ID=$1
ssh_port=$2
password=$3


TIMEOUT=60
WAIT_TIME=5
ELAPSED_TIME=0

# wait for the VM to boot
log_debug $vm_ID "Waiting for Linux VM '$vm_ID' to boot..."
while true; do
    # check if the VM is running
    VM_STATE=$(virsh domstate "$vm_ID" 2>&1)
    if [ "$VM_STATE" == "running" ]; then
        log_debug $vm_ID "Linux VM '$vm_ID' is running."
        break
    fi

    # check if the timeout has been reached
    if [ $ELAPSED_TIME -ge $TIMEOUT ]; then
        log_debug $vm_ID "Timeout waiting for Linux VM '$vm_ID' to start after $TIMEOUT seconds."
        die "Timeout waiting for Linux VM '$vm_ID' to start after $TIMEOUT seconds."
    fi

    sleep $WAIT_TIME
    ELAPSED_TIME=$((ELAPSED_TIME + WAIT_TIME))
done

# wait for the linux guest agent to start
log_debug $vm_ID "Waiting for Linux VM '$vm_ID' to start the guest agent..."
while true; do
    # check if the guest agent is running
    virsh qemu-agent-command "$vm_ID" '{"execute":"guest-ping"}'
    if [ $? -eq 0 ]; then
        log_debug $vm_ID "Linux VM '$vm_ID' has started the guest agent."
        break
    fi

    # check if the timeout has been reached
    if [ $ELAPSED_TIME -ge $TIMEOUT ]; then
        log_debug $vm_ID "Timeout waiting for Linux VM '$vm_ID' to start the guest agent after $TIMEOUT seconds."
        die "Timeout waiting for Linux VM '$vm_ID' to start the guest agent after $TIMEOUT seconds."
    fi

    sleep $WAIT_TIME
    ELAPSED_TIME=$((ELAPSED_TIME + WAIT_TIME))
done

SSH_SCRIPT="sed -i 's/^#\?Port [0-9]\+/Port ${ssh_port}/' /etc/ssh/sshd_config && systemctl restart sshd"

ELAPSED_TIME=0
SSH_SUCEED_TIMES=0
log_debug $vm_ID "Executing script to change SSH port..."
while true; do
    OUTPUT=$(virsh qemu-agent-command "$vm_ID" '{"execute":"guest-exec","arguments":{"path":"/bin/sh","arg":["-c","'"$SSH_SCRIPT"'"],"capture-output":true}}')
    if [ -n "${OUTPUT}" ]; then
        log_debug $vm_ID "$vm_ID exec bash: $OUTPUT"
        QA_RS=$(jq -r '.return' <<< $OUTPUT)
        if [ -n "${QA_RS}" ]; then
            SSH_SUCEED_TIMES=$((SSH_SUCEED_TIMES + 1))
            if [ ${SSH_SUCEED_TIMES} -gt 2 ]; then
                log_debug $vm_ID "$vm_ID exec script succeed"
                break
            fi
        fi
    fi
    if [ $ELAPSED_TIME -ge $TIMEOUT ]; then
        log_debug $vm_ID "$vm_ID timeout while waiting guest agent ready"
        die "Timeout waiting for Linux VM '$vm_ID' to execute script after $TIMEOUT seconds."
    fi
    sleep $WAIT_TIME
    ELAPSED_TIME=$((ELAPSED_TIME + WAIT_TIME))
done

# change the password
if [ -n "$password" ]; then
    virsh set-user-password --domain inst-$vm_ID --user root --password $password
    [ $? -ne 0 ] && die "Failed to set user password"
    echo "|:-COMMAND-:| $(basename $0) '$1' 'success'"
fi