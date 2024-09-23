# Goction

<p align="center">
    <img src="./goction.png" alt="Logo" width="350">
</p>

Goction is a lightweight and extensible platform designed to create, manage, and execute custom actions (called "goctions") via a command-line interface (CLI), an HTTP API, and a web-based dashboard. It provides powerful tools for automation, integration, and workflow management.

## Table of Contents

1. [Features](#features)
2. [Prerequisites](#prerequisites)
3. [Installation](#installation)
4. [Configuration](#configuration)
5. [Usage](#usage)
   - [Managing Goctions](#managing-goctions)
   - [Service Management](#service-management)
   - [Systemd Service Management](#systemd-service-management)
   - [Using the API](#using-the-api)
   - [Dashboard](#dashboard)
   - [Advanced Features](#advanced-features)
6. [Goctions](#goctions)
   - [Goction Example](#goction-example)
   - [Goction Structure and Creation](#goction-structure-and-creation)
7. [Project Structure](#project-structure)
8. [Security](#security)
9. [Logging](#logging)
10. [Troubleshooting](#troubleshooting)
11. [Uninstallation](#uninstallation)
12. [Contributing](#contributing)
13. [License](#license)

## Features

- Easy creation and management of goctions in Go
- Dynamic loading of goctions via Go plugins
- Intuitive CLI interface
- Secure HTTP API for remote execution
- Web-based dashboard for monitoring and management
- Flexible configuration via JSON
- Advanced logging with logrus
- Integration with systemd for robust service management
- Import and export functionality for easy sharing and backup of goctions

## Prerequisites

- Go 1.21 or higher
- ~~Operating system compatible with systemd (e.g., most Linux distributions)~~
- setfacl _sudo apt install acl_ 

## Installation

### Quick Installation (Recommended)

```bash
curl -sSL https://raw.githubusercontent.com/goction/goction/master/hack/install.sh | sudo bash
```

### Manual Installation

1. Clone the repository:

   ```bash
   git clone https://github.com/goction/goction
   cd goction
   ```

2. Run the installation script:
   ```bash
   sudo ./hack/install.sh
   ```

## Configuration

The JSON configuration file is automatically created during installation:

- Location: `/etc/goction/config.json`

This file contains important settings such as:

- `goctions_dir`: Directory where goctions are stored (`/etc/goction/goctions`)
- `port`: The port number for the HTTP API and dashboard (default: 8080)
- `log_file`: Location of the log file (`/var/log/goction/goction.log`)
- `api_token`: The secret token for API authentication
- `stats_file`: Location of the statistics file (`/var/log/goction/goction_stats.json`)
- `dashboard_username`: Username for dashboard access
- `dashboard_password`: Password for dashboard access

You can modify this file to change these settings. To view or reset the configuration:

```bash
goction config view
goction config reset
```

## Usage

### Managing Goctions

Create a new goction:

```bash
goction new my_goction
```

List all goctions:

```bash
goction list
```

Update a goction:

```bash
goction update my_goction
```

### Service Management

Start the Goction service:

```bash
goction start
```

Stop the Goction service:

```bash
goction stop
```

Or check pkill:

```bash
sudo pkill -f "goction serve"
```

### Systemd Service Management

Goction runs as a background service managed by systemd. Use standard systemd commands:

```bash
sudo systemctl start goction
sudo systemctl stop goction
sudo systemctl restart goction
sudo systemctl status goction
sudo systemctl enable goction
sudo systemctl disable goction
sudo journalctl -u goction
```

### Using the API

Execute a goction via the HTTP API:

```bash
curl -X POST -H "Content-Type: application/json" -H "X-API-Token: your-secret-token" -d '{"args":["arg1", "arg2"]}' http://localhost:8080/api/goctions/my_goction
```

### Dashboard

Access the web-based dashboard:

1. Ensure the Goction service is running.
2. Open your web browser and navigate to `http://localhost:8080` (or the configured address).
3. Log in using the credentials set in your Goction configuration.

The dashboard offers:

- Overview of Goction configuration
- Detailed statistics for each goction
- Execution history
- Real-time logs visualization
- Dark UI for comfortable use

### Advanced Features

Export a goction:

```bash
goction export my_goction
```

Import a goction:

```bash
goction import my_goction.zip
```

View goction statistics:

```bash
goction stats my_goction
```

View recent logs:

```bash
goction logs
```

## Goction Example

Here's an example of a simple goction:

```go
package main

import (
    "encoding/json"
    "fmt"
)

func ExampleGoction(args ...string) (string, error) {
    result := fmt.Sprintf("ExampleGoction executed with %d arguments", len(args))
    response := map[string]string{
        "result": result,
        "action": "example_goction",
    }
    jsonResponse, err := json.Marshal(response)
    if err != nil {
        return "", fmt.Errorf("error creating JSON response: %v", err)
    }
    return string(jsonResponse), nil
}
```

## Goction Structure and Creation

A goction is a Go plugin that follows a specific structure. Here's what you need to know about creating and structuring goctions:

### Goction Structure

Each goction should be in its own directory under `/etc/goction/goctions/` and contain at least two files:

1. `main.go`: This file contains the main logic of your goction.
2. `go.mod`: This file declares the module and its dependencies.

Here's an example of the directory structure for a goction named `my_goction`:

```
/etc/goction/goctions/
└── my_goction/
    ├── main.go
    └── go.mod
```

### Creating a Goction

To create a new goction:

1. Use the `goction new` command:

   ```bash
   goction new my_goction
   ```

   This will create a new directory with a template `main.go` and `go.mod` file.

2. Edit the `main.go` file to implement your goction logic. The main function should have the following signature:

   ```go
   func MyGoction(args ...string) (string, error)
   ```

   Replace `MyGoction` with the actual name of your goction (it should start with an uppercase letter).

3. If your goction requires additional dependencies, add them to the `go.mod` file.

4. Build your goction:

   ```bash
   goction update my_goction
   ```

   This command compiles your goction into a Go plugin (.so file).

### Goction Guidelines

- The main function of your goction should be exported (start with an uppercase letter).
- Goctions can accept any number of string arguments.
- The return value should be a string (often JSON-encoded) and an error.
- Keep your goctions modular and focused on a specific task.
- Use proper error handling within your goctions.
- Document your goction's purpose, inputs, and outputs in comments.

### Example Goction Structure

Here's a more detailed example of a goction structure:

```go
package main

import (
    "encoding/json"
    "fmt"
    "strings"
)

// Concatenate joins all input strings and returns them as a JSON object
func Concatenate(args ...string) (string, error) {
    if len(args) == 0 {
        return "", fmt.Errorf("no arguments provided")
    }

    result := strings.Join(args, " ")
    response := map[string]string{
        "result": result,
        "action": "concatenate",
    }

    jsonResponse, err := json.Marshal(response)
    if err != nil {
        return "", fmt.Errorf("error creating JSON response: %v", err)
    }

    return string(jsonResponse), nil
}
```

This goction concatenates all input strings and returns the result as a JSON object.

Remember to test your goctions thoroughly before deploying them in your Goction environment.

## Project Structure

```
goction/
├── cmd/
│   └── goction/
│       └── main.go
├── internal/
│   ├── api/
│   │   ├── server.go
│   │   └── dashboard/
│   │       ├── dashboard.go
│   │       └── templates/
│   │           ├── dashboard.qtpl
│   │           ├── dashboard.qtpl.go
│   │           ├── login.qtpl
│   │           └── login.qtpl.go
│   ├── cmd/
│   │   ├── commands.go
│   │   ├── search.go
│   │   ├── most_used.go
│   │   └── export_import.go
│   ├── config/
│   │   └── config.go
│   └── stats/
│       └── stats.go
├── pkg/
│   └── goctionutil/
│       └── goctionutil.go
├── hack/
│   ├── install.sh
│   └── remove.sh
├── go.mod
├── go.sum
├── README.md
└── goction.service
```

## Security

Goction uses an API token for API requests and a username/password for dashboard access. To display your current API token:

```bash
goction token
```

Keep these credentials confidential and change them regularly.

## Logging

Logs are written to `/var/log/goction/goction.log`. View them via the dashboard, the `goction logs` command, or `sudo journalctl -u goction`.

## Troubleshooting

If you encounter issues:

1. Check `/var/log/goction/goction.log` for error messages.
2. Ensure all goctions are properly compiled using `goction update`.
3. Verify the Goction service is running with `sudo systemctl status goction`.
4. Check firewall settings for API or dashboard access issues.
5. Use `goction stats` to check execution history of specific goctions.
6. For dashboard issues, verify the port and credentials in `/etc/goction/config.json`.

## Uninstallation

To uninstall Goction:

```bash
sudo ./hack/remove.sh
```

This removes all Goction files, including goctions and configurations. Backup important goctions before uninstalling.

## Contributing

Contributions are welcome! Open an issue or submit a pull request on our GitHub repository. See our [Contributing Guide](CONTRIBUTING.md) for more information.

## License

This project is licensed under the MIT License. See the `LICENSE` file for details.
