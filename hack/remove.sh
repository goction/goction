#!/bin/bash

set -e

# Terminal colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Default installation directory
INSTALL_DIR="/opt/goction"

# Function to display messages
print_message() {
    echo -e "${GREEN}[Goction Uninstaller] ${1}${NC}"
}

print_error() {
    echo -e "${RED}[Error] ${1}${NC}"
}

print_warning() {
    echo -e "${YELLOW}[Warning] ${1}${NC}"
}

# Check if the user is root
if [ "$EUID" -ne 0 ]; then
    print_error "This script must be run as root"
    exit 1
fi

# Function to check for running Goction processes
check_running_processes() {
    if pgrep -f goction > /dev/null; then
        print_error "Goction processes are still running. Please stop them before uninstalling."
        print_message "You can stop the Goction service with: sudo systemctl stop goction"
        exit 1
    fi
}

# Function to remove system files and configurations
remove_system_files() {
    print_message "Removing Goction system files and configurations..."
    
    # Remove systemd service file
    if [ -f /etc/systemd/system/goction.service ]; then
        rm -f /etc/systemd/system/goction.service
        systemctl daemon-reload
    fi
    
    # Remove configuration files
    rm -rf /etc/goction

    # Remove binary
    rm -f /usr/local/bin/goction

    # Remove Goction user and home directory
    if id "goction" &>/dev/null; then
        userdel -r goction 2>/dev/null || true
    fi

    # Remove any remaining Goction-related files
    find /home -name ".goction*" -delete
    find /root -name ".goction*" -delete
    find /tmp -name "goction*" -delete

    print_message "System files and configurations removed."
}

# Function to remove installation directory
remove_install_dir() {
    if [ -d "$INSTALL_DIR" ]; then
        print_warning "The installation directory $INSTALL_DIR exists."
        read -p "Do you want to remove this directory? (y/N): " remove_dir
        if [[ $remove_dir =~ ^[Yy]$ ]]; then
            print_message "Removing installation directory..."
            rm -rf "$INSTALL_DIR"
            print_message "Installation directory removed."
        else
            print_message "Installation directory will be kept."
        fi
    else
        print_message "Installation directory $INSTALL_DIR does not exist."
    fi
}

# Function to remove personal goctions
remove_personal_goctions() {
    # Get the actual username (not root)
    local ACTUAL_USER=$(logname 2>/dev/null || echo $SUDO_USER)
    local USER_HOME=$(getent passwd $ACTUAL_USER | cut -d: -f6)
    local GOCTIONS_DIR="$USER_HOME/.config/goction/goctions"

    if [ -d "$GOCTIONS_DIR" ]; then
        print_warning "Personal goctions directory found at $GOCTIONS_DIR"
        read -p "Do you want to remove all personal goctions? (y/N): " remove_goctions
        if [[ $remove_goctions =~ ^[Yy]$ ]]; then
            print_message "Removing personal goctions..."
            rm -rf "$GOCTIONS_DIR"
            print_message "Personal goctions removed."
        else
            print_message "Personal goctions will be kept."
        fi
    else
        print_message "No personal goctions directory found at $GOCTIONS_DIR"
    fi
}

# Confirmation before uninstallation
print_warning "This will uninstall Goction and remove all associated files and configurations."
read -p "Are you sure you want to proceed? (y/N): " confirm
if [[ $confirm != [yY] && $confirm != [yY][eE][sS] ]]; then
    print_message "Uninstallation cancelled."
    exit 0
fi

# Check for running processes
check_running_processes

# Perform uninstallation
print_message "Starting Goction uninstallation..."

# Stop Goction service if it's running
if systemctl is-active --quiet goction.service; then
    print_message "Stopping Goction service..."
    systemctl stop goction.service
fi

# Disable Goction service
print_message "Disabling Goction service..."
systemctl disable goction.service 2>/dev/null || true

# Remove system files and configurations
remove_system_files

# Ask about removing the installation directory
remove_install_dir

# Ask about removing personal goctions
remove_personal_goctions

print_message "Goction has been successfully uninstalled from your system."
print_warning "If you have any custom data or configurations outside of the standard locations, you may need to remove them manually."

exit 0