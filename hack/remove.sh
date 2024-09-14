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

# Stop and disable the Goction service
print_message "Stopping and disabling Goction service..."
systemctl stop goction.service
systemctl disable goction.service

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
read -p "Enter the Goction installation path [/opt/goction]: " INSTALL_PATH
INSTALL_PATH=${INSTALL_PATH:-/opt/goction}
rm -rf $INSTALL_PATH

# Remove the Goction user and home directory
print_message "Removing Goction user and home directory..."
userdel -r goction 2>/dev/null || true

# Remove Goction configuration
print_message "Removing Goction configuration..."
rm -rf /etc/goction
rm -rf /home/goction/.config/goction

# Final cleanup
print_message "Performing final cleanup..."
rm -rf /home/goction

print_message "Goction has been successfully uninstalled from your system."
print_warning "If you have any custom data or configurations outside of the standard locations, you may need to remove them manually."