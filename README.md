# Goction

<p align="center">
    <img src="./goction.png" alt="Logo" width="350">
</p>

Goction is a lightweight and extensible framework designed to create, manage, and execute custom actions (called "goctions") via a command-line interface (CLI) and an HTTP API. It also provides a dashboard for monitoring and managing goctions.

## Table of Contents

1. [Features](#features)
2. [Prerequisites](#prerequisites)
3. [Installation](#installation)
4. [Configuration](#configuration)
5. [Usage](#usage)
   - [Managing Goctions](#managing-goctions)
   - [Service Management](#service-management)
   - [Using the API](#using-the-api)
   - [Dashboard and Execution](#dashboard-and-execution)
   - [Advanced Features](#advanced-features)
6. [Goction Example](#goction-example)
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
- Console-based monitoring dashboard
- Flexible configuration via JSON
- Advanced logging with logrus
- Integration with systemd for simple service management
- Search functionality for goctions
- Most used goctions tracking
- Export and import goctions for easy sharing and backup
- Comprehensive statistics and execution history

## Prerequisites

- Go 1.16 or higher
- Operating system compatible with systemd (e.g., most Linux distributions)

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

The JSON configuration file is automatically created on first launch:

- For root: `/etc/goction/config.json`
- For other users: `~/.config/goction/config.json`

You can modify this file to change settings such as the port number, log file location, or API token.

To view or reset the configuration:

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

Search for goctions:

```bash
goction search <query>
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

### Using the API

Execute a goction via the HTTP API:

```bash
curl -X POST -H "Content-Type: application/json" -H "X-API-Token: your-secret-token" -d '{"args":["arg1", "arg2"]}' http://localhost:8080/goctions/my_goction
```

### Dashboard and Execution

Display the dashboard:

```bash
goction dashboard
```

Execute a goction:

```bash
goction run my_goction [arg1 arg2 ...]
```

### Advanced Features

Show most used goctions:

```bash
goction most-used
```

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

## Project Structure

```
goction/
├── cmd/
│   └── goction/
│       └── main.go
├── internal/
│   ├── api/
│   │   └── server.go
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

Goction uses an API token to secure requests. To display your current API token:

```bash
goction token
```

Ensure you keep this token confidential and change it regularly.

## Logging

Logs are managed by logrus and are written to the file specified in the configuration. They can be viewed via the dashboard or by using the `goction logs` command.

## Troubleshooting

If you encounter issues:

1. Check the log file for error messages.
2. Ensure all goctions are properly compiled using the `goction update` command.
3. Verify that the Goction service is running using `goction start`.
4. Check your firewall settings if you're having trouble with the API.
5. Use the `goction stats` command to check the execution history of a specific goction.

## Uninstallation

To uninstall Goction, use the removal script:

```bash
sudo ./hack/remove.sh
```

This script will remove all Goction files, including your goctions and configuration. Make sure to backup any important goctions before uninstalling.

## Contributing

Contributions are welcome! Feel free to open an issue or submit a pull request on our GitHub repository. For more information, see our [Contributing Guide](CONTRIBUTING.md).

## License

This project is licensed under the MIT License. See the `LICENSE` file for more details.
