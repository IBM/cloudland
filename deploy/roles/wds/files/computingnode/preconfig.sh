#!/bin/bash

# Define log file
LOG_FILE="/tmp/preinstall_mlnx.log"

# Record the initial directory
INITIAL_DIR=$(pwd)

# Define a function to log messages
log() {
    echo "$(date '+%Y-%m-%d %H:%M:%S') - $1" | tee -a "$LOG_FILE"
}

# Function to check the status of the last command and exit if it fails
check_status() {
    if [ $? -ne 0 ]; then
        log "Error encountered. Exiting."
        exit 1
    fi
}

# Function to change directories safely
safe_cd() {
    cd "$INITIAL_DIR"
    cd "$1" &>> "$LOG_FILE"
    check_status
}

# Log config
log "Starting config process..."

# Go back to main folder
log "Changing directory back to $INITIAL_DIR..."
safe_cd "$INITIAL_DIR"

# Change directory
log "Changing directory to PEG_Ubuntu_22.04_mlnxDriver_deps/..."
safe_cd "PEG_Ubuntu_22.04_mlnxDriver_deps/"

# Set up rc.local
log "Setting up rc.local..."
sudo cp rc.local /etc/rc.local &>> "$LOG_FILE"
sudo chmod +x /etc/rc.local &>> "$LOG_FILE"
sudo cp rc-local.service /lib/systemd/system/rc-local.service &>> "$LOG_FILE"
sudo mkdir -p /etc/rc.d/ && sudo ln -sf /etc/rc.local /etc/rc.d/rc.local
sudo systemctl daemon-reload &>> "$LOG_FILE"
sudo systemctl enable rc-local &>> "$LOG_FILE"
sudo systemctl start rc-local.service &>> "$LOG_FILE"
check_status

# Check rc-local service status
log "Checking rc-local service status..."
sudo systemctl status rc-local.service &>> "$LOG_FILE"
check_status

# Reload nvme modules
log "Reloading NVMe modules..."
sudo rmmod nvme &>> "$LOG_FILE"
sudo rmmod nvme-core &>> "$LOG_FILE"
sudo modprobe nvme && sudo modprobe nvme-rdma &>> "$LOG_FILE"
check_status

# List NVMe modules
log "Listing NVMe modules..."
sudo lsmod | grep nvme &>> "$LOG_FILE"
check_status


if [ "$1" == "computingnode" ]; then
    # modify os configure for computing node.
    log "Modify os configure for computing node...."
    sudo ln -sf /usr/bin/bash /bin/sh
    sudo systemctl stop chrony
    sudo ln -sf /etc/chrony/chrony.conf /etc/chrony.conf
    sudo rm -f /etc/systemd/system/chronyd.service
    sudo mkdir -p /etc/systemd/system/multiuser.target.wants/
    sudo cp /lib/systemd/system/chrony.service /lib/systemd/system/chronyd.service
    sudo ln -sf /lib/systemd/system/chronyd.service /etc/systemd/system/multiuser.target.wants/chronyd.service
    sudo rm -f /etc/systemd/system/chronyd.service
    sudo systemctl daemon-reload
    sudo systemctl start chronyd

    sudo sed -i "s/#PermitRootLogin prohibit-password/PermitRootLogin yes/g" /etc/ssh/sshd_config
    sudo cat /etc/ssh/sshd_config | grep "PermitRootLogin yes" &>> "$LOG_FILE"
    sudo systemctl restart sshd

    sudo echo 'root:root123' | sudo chpasswd root
fi

log "Pre-configuration completed successfully."
