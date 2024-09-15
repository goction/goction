package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"goction/internal/api"
	"goction/internal/cmd"
	"goction/internal/config"
	"goction/internal/stats"

	"github.com/sirupsen/logrus"
)

func main() {
	fmt.Println("Starting Goction...")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Error loading configuration: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	logger, err := initializeLogger(cfg)
	if err != nil {
		fmt.Printf("Error initializing logger: %v\n", err)
		os.Exit(1)
	}

	// Initialize stats manager
	statsManager, err := stats.NewManager(cfg.StatsFile)
	if err != nil {
		logger.Fatalf("Failed to create stats manager: %v", err)
	}

	// Check command-line arguments
	if len(os.Args) < 2 {
		fmt.Println("Usage: goction [new|start|stop|serve|list|update|token|stats|dashboard|run|export|import|config|logs|self-update]")
		os.Exit(1)
	}

	// Execute the appropriate command
	command := os.Args[1]
	args := os.Args[2:]

	logger.Infof("Executing command: %s", command)

	err = executeCommand(command, args, cfg, statsManager, logger)

	if err != nil {
		logger.Fatalf("Error executing command: %v", err)
	}

	fmt.Println("Goction execution completed.")
}

func initializeLogger(cfg *config.Config) (*logrus.Logger, error) {
	logger := logrus.New()

	// Ensure log directory exists
	logDir := filepath.Dir(cfg.LogFile)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	// Set up file logging
	file, err := os.OpenFile(cfg.LogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	// Use multi-writer to log to both file and console
	mw := io.MultiWriter(os.Stdout, file)
	logger.SetOutput(mw)

	logger.SetLevel(logrus.DebugLevel)
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})

	return logger, nil
}

func executeCommand(command string, args []string, cfg *config.Config, statsManager *stats.Manager, logger *logrus.Logger) error {
	switch command {
	case "new":
		return cmd.CreateNewGoction(args, cfg)
	case "start":
		return cmd.StartService(cfg)
	case "stop":
		return cmd.StopService(cfg)
	case "serve":
		return serveAPI(cfg, logger)
	case "list":
		return cmd.ListGoctions(cfg)
	case "update":
		return cmd.UpdateGoction(args, cfg)
	case "token":
		return cmd.ShowToken(cfg)
	case "stats":
		return cmd.ShowStats(args, statsManager)
	case "dashboard":
		return cmd.ShowDashboard(cfg)
	case "run":
		if len(args) < 1 {
			return fmt.Errorf("Usage: goction run <goction-name> [arg1 arg2 ...]")
		}
		return cmd.RunGoction(args[0], args[1:], cfg)
	case "config":
		if len(args) == 0 {
			return fmt.Errorf("Usage: goction config [view|reset]")
		}
		switch args[0] {
		case "view":
			return cmd.ConfigView(cfg)
		case "reset":
			return cmd.ConfigReset(cfg)
		default:
			return fmt.Errorf("Unknown config subcommand: %s", args[0])
		}
	case "logs":
		return cmd.ShowLogs(cfg)
	case "self-update":
		return cmd.SelfUpdate()
	case "export":
		return cmd.ExportGoction(args, cfg)
	case "import":
		return cmd.ImportGoction(args, cfg)
	default:
		return fmt.Errorf("Unknown command: %s", command)
	}
}

func serveAPI(cfg *config.Config, logger *logrus.Logger) error {
	fmt.Println("Initializing server...")
	server, err := api.NewServer(cfg)
	if err != nil {
		return fmt.Errorf("Failed to create server: %v", err)
	}
	fmt.Println("Starting server...")
	return server.Start()
}
