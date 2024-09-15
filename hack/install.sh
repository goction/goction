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
GOCTION_GROUP="goction"
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

# Function to create Goction user
create_goction_user() {
    print_message "Creating Goction user..."
    if ! getent group "$GOCTION_GROUP" > /dev/null 2>&1 ; then
        groupadd -r "$GOCTION_GROUP"
    fi
    if ! getent passwd "$GOCTION_USER" > /dev/null 2>&1 ; then
        useradd -r -g "$GOCTION_GROUP" -d "/home/$GOCTION_USER" -m -s /bin/false "$GOCTION_USER"
    fi
    log_message "Goction user created"
}

initialize_config() {
    print_message "Initializing configuration..."
    
    CONFIG_FILE="/etc/goction/config.json"
    if [ ! -f "$CONFIG_FILE" ]; then
        cat << EOF > "$CONFIG_FILE"
{
  "goctions_dir": "/etc/goction/goctions",
  "port": 8080,
  "log_file": "/var/log/goction/goction.log",
  "api_token": "$(uuidgen)",
  "stats_file": "/var/log/goction/goction_stats.json",
  "dashboard_username": "admin",
  "dashboard_password": "$(uuidgen)"
}
EOF
    fi
    
    log_message "Configuration initialized"
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

    # Set correct ownership for the installation directory
    chown -R $GOCTION_USER:$GOCTION_GROUP $INSTALL_DIR

    log_message "Goction installed in $INSTALL_DIR"
}

# Function to setup environment
setup_environment() {
    print_message "Setting up environment..."
    
    # Add Go to system-wide PATH
    echo "export PATH=$PATH:/usr/local/go/bin" > /etc/profile.d/goction.sh
    chmod +x /etc/profile.d/goction.sh
    
    # Reload shell to apply changes
    source /etc/profile.d/goction.sh
    
    log_message "Environment setup completed"
}

# Function to setup permissions
setup_permissions() {
    print_message "Setting up permissions..."
    
    # Ensure directories exist
    mkdir -p /etc/goction/goctions
    mkdir -p /var/log/goction
    
    # Initialize configuration if it doesn't exist
    initialize_config
    
    # Set ownership
    chown -R $GOCTION_USER:$GOCTION_GROUP /etc/goction
    chown -R $GOCTION_USER:$GOCTION_GROUP /var/log/goction
    
    # Set permissions
    chmod 755 /etc/goction
    chmod 775 /etc/goction/goctions
    chmod 775 /var/log/goction
    chmod 664 /etc/goction/config.json
    touch /var/log/goction/goction.log
    touch /var/log/goction/goction_stats.json
    chmod 664 /var/log/goction/goction.log
    chmod 664 /var/log/goction/goction_stats.json
    
    # Ensure the goction user has write access
    chown $GOCTION_USER:$GOCTION_GROUP /var/log/goction/goction.log
    chown $GOCTION_USER:$GOCTION_GROUP /var/log/goction/goction_stats.json
    
    # Add current user to goction group
    usermod -aG $GOCTION_GROUP $SUDO_USER
    
    log_message "Permissions set up completed"
}

initialize_stats() {
    print_message "Initializing stats file..."
    
    STATS_FILE="/var/log/goction/goction_stats.json"
    if [ ! -s "$STATS_FILE" ]; then
        echo '{"stats":{},"history":{}}' > "$STATS_FILE"
    fi
    
    log_message "Stats file initialized"
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
Group=$GOCTION_GROUP
WorkingDirectory=/etc/goction

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
    setup_environment
    setup_permissions
    create_systemd_service

    systemctl start goction.service
    print_message "Goction service started"

    print_message "Goction has been successfully installed!"
    print_message "You can now use the 'goction' command to manage your goctions."
    print_message "Goction is running on port $GOCTION_PORT"
    print_message "Dashboard credentials can be found in /etc/goction/config.json"
    log_message "Installation completed successfully"
}

# Run the main installation process
main

exit 0