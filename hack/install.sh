#!/bin/bash

set -e

# Terminal colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Default values
INSTALL_DIR="/opt/goction"
GOCTION_USER="goction"
GOCTION_PORT=8080

# Function to display messages
print_message() {
    echo -e "${GREEN}[Goction Installer] ${1}${NC}"
}

print_error() {
    echo -e "${RED}[Error] ${1}${NC}" >&2
}

print_warning() {
    echo -e "${YELLOW}[Warning] ${1}${NC}"
}

print_debug() {
    echo -e "${YELLOW}[Debug] ${1}${NC}"
}

# Function to log messages
log_message() {
    echo "$(date): $1" >> /var/log/goction_install.log
}

# Error handling
trap 'print_error "An error occurred. Exiting..."; log_message "Installation failed"; exit 1' ERR

# Check if the user is root
if [ "$EUID" -ne 0 ]; then
    print_error "This script must be run as root"
    exit 1
fi

# Ensure SUDO_USER is set
if [ -z "$SUDO_USER" ]; then
    print_error "This script must be run with sudo"
    exit 1
fi

# Preserve the user's environment
USER_HOME=$(getent passwd $SUDO_USER | cut -d: -f6)
USER_PATH=$(sudo -u $SUDO_USER bash -c 'echo $PATH')
export PATH=$PATH:$USER_PATH
print_debug "User PATH: $USER_PATH"
print_debug "Current PATH: $PATH"

# Function to check system dependencies
check_dependencies() {
    print_message "Checking system dependencies..."
    
    # Function to find command, considering both root and normal user environments
    find_command() {
        local cmd=$1
        local cmd_path=""
        
        # Check in current PATH
        cmd_path=$(which $cmd 2>/dev/null)
        if [ -n "$cmd_path" ]; then
            echo $cmd_path
            return 0
        fi
        
        # Check in user's PATH
        cmd_path=$(sudo -u $SUDO_USER which $cmd 2>/dev/null)
        if [ -n "$cmd_path" ]; then
            echo $cmd_path
            return 0
        fi
        
        # Check common Go installation directories
        for dir in "/usr/local/go/bin" "$USER_HOME/go/bin" "$USER_HOME/.go/bin" "/snap/bin"; do
            if [ -x "$dir/$cmd" ]; then
                echo "$dir/$cmd"
                return 0
            fi
        done
        
        return 1
    }

    for cmd in git curl; do
        cmd_path=$(find_command $cmd)
        if [ -z "$cmd_path" ]; then
            print_error "$cmd is not installed or not in PATH. Please install it and try again."
            exit 1
        else
            print_debug "Found $cmd at: $cmd_path"
        fi
    done

    # Special check for Go
    GO_CMD=$(find_command go)
    if [ -z "$GO_CMD" ]; then
        print_error "Go is not installed or not in PATH. Please install it and try again."
        print_debug "Searched in PATH: $PATH"
        print_debug "Searched in user's PATH: $USER_PATH"
        exit 1
    else
        print_debug "Found Go at: $GO_CMD"
    fi

    # Check Go version
    GO_VERSION=$($GO_CMD version 2>&1 | awk '{print $3}' | sed 's/go//')
    MIN_VERSION="1.16"
    if [ "$(printf '%s\n' "$MIN_VERSION" "$GO_VERSION" | sort -V | head -n1)" != "$MIN_VERSION" ]; then 
        print_error "Go version $MIN_VERSION or higher is required. You have $GO_VERSION."
        exit 1
    fi

    print_message "Go version $GO_VERSION found at $GO_CMD"
    log_message "System dependencies checked"
}

# Function to update PATH
update_path() {
    print_message "Updating system PATH..."
    echo "export PATH=\$PATH:$INSTALL_DIR" > /etc/profile.d/goction.sh
    chmod +x /etc/profile.d/goction.sh
    source /etc/profile.d/goction.sh
    log_message "System PATH updated"
}

# Function to configure firewall
configure_firewall() {
    print_message "Configuring firewall..."
    if command -v ufw >/dev/null 2>&1; then
        ufw allow $GOCTION_PORT/tcp
        print_message "UFW firewall rule added for port $GOCTION_PORT"
    elif command -v firewall-cmd >/dev/null 2>&1; then
        firewall-cmd --permanent --add-port=$GOCTION_PORT/tcp
        firewall-cmd --reload
        print_message "FirewallD rule added for port $GOCTION_PORT"
    else
        print_warning "Unable to configure firewall automatically. Please ensure port $GOCTION_PORT is open."
    fi
    log_message "Firewall configured for port $GOCTION_PORT"
}

# Function to create Goction user
create_goction_user() {
    print_message "Creating Goction user..."
    useradd -r -s /bin/false $GOCTION_USER
    mkdir -p /home/$GOCTION_USER
    chown $GOCTION_USER:$GOCTION_USER /home/$GOCTION_USER
    log_message "Goction user created"
}

# Function to install Goction
install_goction() {
    print_message "Installing Goction..."

    if [ -d "$INSTALL_DIR" ]; then
        print_warning "Installation directory $INSTALL_DIR already exists."
        read -p "Do you want to remove the existing directory and continue? (y/N): " remove_dir
        if [[ $remove_dir =~ ^[Yy]$ ]]; then
            print_message "Removing existing directory..."
            rm -rf "$INSTALL_DIR"
        else
            print_error "Installation cancelled by user."
            exit 1
        fi
    fi

    mkdir -p $INSTALL_DIR
    git clone https://github.com/goction/goction.git $INSTALL_DIR
    cd $INSTALL_DIR

    # Ensure the current user has write permissions in the install directory
    chown -R $SUDO_USER:$SUDO_USER $INSTALL_DIR

    # Build Goction as the original user
    sudo -u $SUDO_USER $GO_CMD build -o goction cmd/goction/main.go

    # Copy the binary to /usr/local/bin and set correct permissions
    cp goction /usr/local/bin/
    chown root:root /usr/local/bin/goction
    chmod 755 /usr/local/bin/goction

    # Set up dashboard credentials
    read -p "Enter a username for the dashboard: " dashboard_username
    read -s -p "Enter a password for the dashboard: " dashboard_password
    echo

    # Update config file with dashboard credentials
    config_file="~/.config/goction/config.json"
    if [ -f "$config_file" ]; then
        sed -i "s/\"dashboard_username\": \"\"/\"dashboard_username\": \"$dashboard_username\"/" "$config_file"
        sed -i "s/\"dashboard_password\": \"\"/\"dashboard_password\": \"$dashboard_password\"/" "$config_file"
    else
        print_warning "Config file not found. Dashboard credentials will need to be set manually."
    fi

    # Set correct ownership for the installation directory
    chown -R $GOCTION_USER:$GOCTION_USER $INSTALL_DIR

    log_message "Goction installed in $INSTALL_DIR"
}

# Function to create systemd service
create_systemd_service() {
    print_message "Creating systemd service..."
    cat << EOF > /etc/systemd/system/goction.service
[Unit]
Description=Goction API Service
After=network.target

[Service]
ExecStart=/usr/local/bin/goction serve
Restart=on-failure
User=$GOCTION_USER
Group=$GOCTION_USER
Environment=PATH=/usr/bin:/usr/local/bin:$PATH
WorkingDirectory=$INSTALL_DIR

[Install]
WantedBy=multi-user.target
EOF
    systemctl daemon-reload
    systemctl enable goction.service
    log_message "Systemd service created and enabled"
}

# Main installation process
main() {
    print_message "Starting Goction installation..."
    log_message "Installation started"

    check_dependencies

    create_goction_user

    install_goction

    create_systemd_service

    update_path

    read -p "Do you want to configure the firewall? (y/N): " configure_fw
    if [[ $configure_fw =~ ^[Yy]$ ]]; then
        configure_firewall
    fi

    systemctl start goction.service
    print_message "Goction service started"

    print_message "Goction has been successfully installed!"
    print_message "You can now use the 'goction' command to manage your goctions."
    print_message "Goction is running on port $GOCTION_PORT"
    log_message "Installation completed successfully"
}

# Run the main installation process
main

exit 0