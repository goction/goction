#!/bin/bash

set -e

# Terminal colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to display messages
print_message() {
    echo -e "${GREEN}[Goction Installer] ${1}${NC}"
}

print_error() {
    echo -e "${RED}[Error] ${1}${NC}"
}

print_warning() {
    echo -e "${YELLOW}[Warning] ${1}${NC}"
}

# Preserve the user's environment
if [ "$SUDO_USER" ]; then
    USER_HOME=$(getent passwd $SUDO_USER | cut -d: -f6)
    export PATH=$PATH:$(sudo -u $SUDO_USER bash -c 'echo $PATH')
    export GOPATH=$(sudo -u $SUDO_USER go env GOPATH)
fi

# Check if Go is installed
if ! command -v go &> /dev/null; then
    print_error "Go is not installed. Please install Go before continuing."
    exit 1
fi

# Check Go version
GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
MIN_VERSION="1.16"
if [ "$(printf '%s\n' "$MIN_VERSION" "$GO_VERSION" | sort -V | head -n1)" != "$MIN_VERSION" ]; then 
    print_error "Go version $MIN_VERSION or higher is required. You have $GO_VERSION."
    exit 1
fi

# Ask for the installation path
read -p "Enter the installation path for Goction [/opt/goction]: " INSTALL_PATH
INSTALL_PATH=${INSTALL_PATH:-/opt/goction}

# Create the installation directory
print_message "Creating installation directory..."
mkdir -p $INSTALL_PATH
cd $INSTALL_PATH

# Clone the Goction repository
print_message "Cloning Goction repository..."
git clone https://github.com/benoitpetit/goction.git .

# Build the project
print_message "Building Goction..."
go build -o goction cmd/goction/main.go

# Copy the executable
print_message "Installing the executable..."
cp goction /usr/local/bin/

# Create goction user
print_message "Creating goction user..."
useradd -r -s /bin/false goction
mkdir -p /home/goction
chown goction:goction /home/goction

# Create systemd service file
print_message "Creating systemd service file..."
cat << EOF > /etc/systemd/system/goction.service
[Unit]
Description=Goction API Service
After=network.target

[Service]
ExecStart=/usr/local/bin/goction serve
Restart=on-failure
User=goction
Group=goction
Environment=PATH=/usr/bin:/usr/local/bin:$PATH
Environment=GOPATH=$GOPATH
WorkingDirectory=/home/goction

[Install]
WantedBy=multi-user.target
EOF

# Reload systemd
print_message "Reloading systemd..."
systemctl daemon-reload

# Enable and start the service
print_message "Enabling and starting Goction service..."
systemctl enable goction.service
systemctl start goction.service

# Check service status
if systemctl is-active --quiet goction.service; then
    print_message "Goction service is running."
else
    print_error "Goction service could not be started. Please check the logs with 'journalctl -u goction.service'"
fi

# Create default goctions directory
print_message "Creating default goctions directory..."
mkdir -p /home/goction/.config/goction/goctions
chown -R goction:goction /home/goction/.config/goction

# Display configuration information
print_message "Installation completed!"
print_message "You can now use the 'goction' command to manage your goctions."
print_message "Use 'goction dashboard' to view the dashboard and get more information."

# Final warning
print_warning "Don't forget to secure your installation by changing default tokens and passwords."