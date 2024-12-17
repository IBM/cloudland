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

# Log start
log "Starting installation process..."

# Extract tar file
PEG_PKG=PEG_Ubuntu_22.04_mlnxDriver_deps
wget https://dev-repo.raksmart.com/packages/WDS/computingnode/${PEG_PKG}.tar.gz
log "Extracting $PEG_PKG..."
tar -xvzf ${PEG_PKG}.tar.gz &>> "$LOG_FILE"
check_status

# Change directory
log "Changing directory to $PEG_PKG/archives/..."
safe_cd "$PEG_PKG/archives/"

# Install .deb packages
log "Installing all .deb packages..."
sudo dpkg -i *.deb &>> "$LOG_FILE"

if [ "$1" == "computingnode" ]; then
    # Change directory
    log "Changing directory to PEG_Ubuntu_22.04_mlnxDriver_deps/archives/computingnode/..."
    safe_cd "$PEG_PKG/archives/computingnode/"

    log "Installing dracut_051-1_all..."
    sudo dpkg -i *.deb &>> "$LOG_FILE"
    check_status
fi

safe_cd "$INITIAL_DIR"
rm -rf ${PEG_PKG}*

# Go back to main folder
log "Changing directory back to $INITIAL_DIR..."
safe_cd "$INITIAL_DIR"

# Extract the Mellanox OFED driver package
OFED_PKG=MLNX_OFED_LINUX-5.8-3.0.7.0-ubuntu22.04-x86_64
wget https://dev-repo.raksmart.com/packages/WDS/computingnode/${OFED_PKG}.tgz
log "Extracting $OFED_PKG..."
tar -xvzf ${OFED_PKG}.tgz &>> "$LOG_FILE"
check_status

# Change directory to MLNX_OFED_LINUX
log "Changing directory to MLNX_OFED_LINUX..."
safe_cd "$OFED_PKG"

# Install MLNX OFED
log "Installing Mellanox OFED drivers, aroud 10 mins..."
sudo ./mlnxofedinstall --with-nvmf --force &>> "$LOG_FILE"
check_status
safe_cd "$INITIAL_DIR"
rm -rf ${OFED_PKG}*

# Rebuild initramfs
log "Rebuilding initramfs..."
sudo dracut -f &>> "$LOG_FILE"
check_status

log "Pre-installation completed successfully."
