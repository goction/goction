#!/bin/bash

set -e

# Terminal colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

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

# Confirm uninstallation
read -p "Are you sure you want to uninstall Goction? This will remove all Goction files and configurations. (y/N) " -n 1 -r

echo
if [[ ! $REPLY =~ ^[Yy]$ ]]
then
    print_message "Uninstallation cancelled."
    exit 0
fi

# Stop and disable the Goction service
print_message "Stopping and disabling Goction service..."
systemctl stop goction.service || print_warning "Failed to stop Goction service. It might not be running."
systemctl disable goction.service || print_warning "Failed to disable Goction service. It might not be enabled."

# Remove the systemd service file
print_message "Removing systemd service file..."
rm -f /etc/systemd/system/goction.service

# Reload systemd
print_message "Reloading systemd..."
systemctl daemon-reload

# Remove the Goction executable
print_message "Removing Goction executable..."
rm -f /usr/local/bin/goction

# Remove the Goction installation directory
print_message "Removing Goction installation directory..."
INSTALL_PATH="/opt/goction"
rm -rf $INSTALL_PATH

# Remove the Goction user and group
print_message "Removing Goction user and group..."
userdel -r goction 2>/dev/null || print_warning "Failed to remove Goction user. It might not exist."
groupdel goction 2>/dev/null || print_warning "Failed to remove Goction group. It might not exist."

# Remove Goction system-wide configuration
print_message "Removing system-wide Goction configuration..."
rm -rf /etc/goction

# Remove Goction logs and stats
print_message "Removing Goction logs and stats..."
rm -rf /var/log/goction

# Clean up any remaining files in /tmp
print_message "Cleaning up temporary files..."
rm -rf /tmp/goction*

print_message "Goction has been successfully uninstalled from your system."
print_warning "If you have any custom data or configurations outside of the standard locations, you may need to remove them manually."