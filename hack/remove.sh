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
read -p "Enter the Goction installation path [/opt/goction]: " INSTALL_PATH
INSTALL_PATH=${INSTALL_PATH:-/opt/goction}
rm -rf $INSTALL_PATH

# Remove the Goction user and home directory
print_message "Removing Goction user and home directory..."
userdel -r goction 2>/dev/null || print_warning "Failed to remove Goction user. It might not exist."

# Remove Goction system-wide configuration
print_message "Removing system-wide Goction configuration..."
rm -rf /etc/goction

# Remove user-specific Goction configuration
if [ "$SUDO_USER" ]; then
    USER_HOME=$(getent passwd $SUDO_USER | cut -d: -f6)
    CONFIG_DIR="$USER_HOME/.config/goction"
    
    if [ -d "$CONFIG_DIR" ]; then
        print_message "Removing user-specific Goction configuration..."
        
        # Check if there are any goctions in the user's directory
        if [ -d "$CONFIG_DIR/goctions" ] && [ "$(ls -A $CONFIG_DIR/goctions)" ]; then
            read -p "Do you want to remove all user goctions in $CONFIG_DIR/goctions? (y/N) " -n 1 -r
            echo
            if [[ $REPLY =~ ^[Yy]$ ]]
            then
                print_message "Removing user goctions..."
                rm -rf "$CONFIG_DIR"
            else
                print_message "Keeping user goctions. Removing other Goction configuration files..."
                find "$CONFIG_DIR" -mindepth 1 -maxdepth 1 ! -name "goctions" -exec rm -rf {} +
            fi
        else
            rm -rf "$CONFIG_DIR"
        fi
    fi
else
    print_warning "Cannot determine user home directory. You may need to manually remove ~/.config/goction if it exists."
fi

print_message "Goction has been successfully uninstalled from your system."
print_warning "If you have any custom data or configurations outside of the standard locations, you may need to remove them manually."