package cmd

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"plugin"
	"runtime"
	"strings"
	"time"

	"goction/internal/config"
	"goction/internal/stats"
	"goction/pkg/goctionutil"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
)

// CreateNewGoction creates a new goction file
func CreateNewGoction(args []string, cfg *config.Config) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: goction new <goction-name>")
	}
	name := args[0]

	goctionDir := filepath.Join(cfg.GoctionsDir, name)
	if err := os.MkdirAll(goctionDir, 0755); err != nil {
		return fmt.Errorf("failed to create goction directory: %w", err)
	}

	filename := filepath.Join(goctionDir, "main.go")
	content := goctionutil.GenerateGoctionTemplate(name)

	if err := os.WriteFile(filename, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write goction file: %w", err)
	}

	// Initialize go.mod
	cmd := exec.Command("go", "mod", "init", name)
	cmd.Dir = goctionDir
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to initialize go.mod: %w", err)
	}

	fmt.Printf("New goction '%s' created successfully\n", name)
	return nil
}

// StartService starts the Goction service
func StartService(cfg *config.Config) error {
	cmd := exec.Command(os.Args[0], "serve")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start service: %w", err)
	}
	fmt.Println("Goction service started in background")
	return nil
}

// StopService stops the Goction service
func StopService(cfg *config.Config) error {
	fmt.Println("Goction service stop command received. Please use system service manager to stop the service.")
	return nil
}

// ListGoctions lists all available goctions
func ListGoctions(cfg *config.Config) error {
	goctions, err := listGoctions(cfg.GoctionsDir)
	if err != nil {
		return fmt.Errorf("failed to list goctions: %w", err)
	}

	fmt.Println("Available Goctions:")
	for _, goction := range goctions {
		fmt.Println("-", goction)
	}
	return nil
}

// UpdateGoction updates an existing goction
func UpdateGoction(args []string, cfg *config.Config) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: goction update <goction-name>")
	}
	name := args[0]

	goctionDir := filepath.Join(cfg.GoctionsDir, name)
	if _, err := os.Stat(goctionDir); os.IsNotExist(err) {
		return fmt.Errorf("goction '%s' does not exist", name)
	}

	cmd := exec.Command("go", "build", "-buildmode=plugin", "-o",
		filepath.Join(goctionDir, name+".so"), ".")
	cmd.Dir = goctionDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to build goction: %w", err)
	}

	fmt.Printf("Goction '%s' updated successfully\n", name)
	return nil
}

// ShowStats displays statistics for goctions
func ShowStats(args []string, statsManager *stats.Manager) error {
	if len(args) == 0 {
		return showAllStats(statsManager)
	}
	return showSpecificStats(args[0], statsManager)
}

func showAllStats(statsManager *stats.Manager) error {
	allStats := statsManager.GetAllStats()
	if len(allStats) == 0 {
		fmt.Println("No statistics available.")
		return nil
	}

	fmt.Println("Goction Statistics:")
	for name, stats := range allStats {
		printGoctionStats(name, stats)
	}
	return nil
}

func showSpecificStats(name string, statsManager *stats.Manager) error {
	stats, ok := statsManager.GetStats(name)
	if !ok {
		return fmt.Errorf("no statistics available for goction '%s'", name)
	}
	printGoctionStats(name, stats)
	return nil
}

// ShowDashboard displays the enhanced dashboard in the command line
func ShowDashboard(cfg *config.Config) error {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#7D56F4")).
		Padding(0, 1)

	sectionStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#7D56F4")).
		MarginTop(1)

	infoStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF"))

	fmt.Println(titleStyle.Render("Goction Dashboard"))
	renderConfigInfo(cfg, sectionStyle, infoStyle)
	renderSystemInfo(sectionStyle, infoStyle)
	renderActionTable(cfg, sectionStyle)
	renderRecentLogs(cfg, sectionStyle, infoStyle)

	return nil
}

// RunGoction executes a specific goction
func RunGoction(name string, args []string, cfg *config.Config) error {
	goctionPath := filepath.Join(cfg.GoctionsDir, name, name+".so")
	plug, err := plugin.Open(goctionPath)
	if err != nil {
		return fmt.Errorf("could not open goction plugin: %w", err)
	}

	sym, err := plug.Lookup(strings.Title(name))
	if err != nil {
		return fmt.Errorf("could not find goction symbol: %w", err)
	}

	goction, ok := sym.(func(...string) (string, error))
	if !ok {
		return fmt.Errorf("unexpected type from module symbol")
	}

	start := time.Now()
	result, err := goction(args...)
	duration := time.Since(start)

	statsManager, err := stats.NewManager(cfg.StatsFile)
	if err != nil {
		return fmt.Errorf("failed to create stats manager: %w", err)
	}
	statsManager.RecordExecution(name, duration, err == nil, result)

	if err != nil {
		return fmt.Errorf("goction execution failed: %w", err)
	}

	fmt.Printf("Goction '%s' executed successfully in %v\n", name, duration)
	fmt.Printf("Result: %s\n", result)

	return nil
}

// ShowToken displays the current API token
func ShowToken(cfg *config.Config) error {
	fmt.Printf("Your current API token is: %s\n", cfg.APIToken)
	return nil
}

// Helper functions

func renderConfigInfo(cfg *config.Config, sectionStyle, infoStyle lipgloss.Style) {
	fmt.Println(sectionStyle.Render("Configuration"))
	fmt.Printf("%s %s\n", infoStyle.Render("API Token:"), cfg.APIToken)
	fmt.Printf("%s %s\n", infoStyle.Render("Goctions Directory:"), cfg.GoctionsDir)
	fmt.Printf("%s %d\n", infoStyle.Render("Server Port:"), cfg.Port)
	fmt.Printf("%s %s\n", infoStyle.Render("Log File:"), cfg.LogFile)
	fmt.Printf("%s %s\n", infoStyle.Render("Stats File:"), cfg.StatsFile)
}

func renderSystemInfo(sectionStyle, infoStyle lipgloss.Style) {
	fmt.Println(sectionStyle.Render("System Information"))
	fmt.Printf("%s %s\n", infoStyle.Render("OS:"), runtime.GOOS)
	fmt.Printf("%s %s\n", infoStyle.Render("Architecture:"), runtime.GOARCH)
	fmt.Printf("%s %d\n", infoStyle.Render("CPUs:"), runtime.NumCPU())

	v, _ := mem.VirtualMemory()
	fmt.Printf("%s %.2f%%\n", infoStyle.Render("Memory Usage:"), v.UsedPercent)

	c, _ := cpu.Percent(time.Second, false)
	fmt.Printf("%s %.2f%%\n", infoStyle.Render("CPU Usage:"), c[0])
}

func renderActionTable(cfg *config.Config, sectionStyle lipgloss.Style) {
	fmt.Println(sectionStyle.Render("Available Actions"))
	actionTable, err := createActionTable(cfg)
	if err != nil {
		fmt.Printf("Error creating action table: %v\n", err)
		return
	}
	fmt.Println(actionTable.View())
}

func renderRecentLogs(cfg *config.Config, sectionStyle, infoStyle lipgloss.Style) {
	fmt.Println(sectionStyle.Render("Recent Logs"))
	logs, err := getRecentLogs(cfg.LogFile, 5)
	if err != nil {
		fmt.Printf("%s Error reading logs: %v\n", infoStyle.Render("•"), err)
	} else {
		for _, log := range logs {
			fmt.Printf("%s %s\n", infoStyle.Render("•"), log)
		}
	}
}

func createActionTable(cfg *config.Config) (table.Model, error) {
	actions, err := listActions(cfg.GoctionsDir)
	if err != nil {
		return table.Model{}, fmt.Errorf("failed to list actions: %w", err)
	}

	statsManager, err := stats.NewManager(cfg.StatsFile)
	if err != nil {
		return table.Model{}, fmt.Errorf("failed to create stats manager: %w", err)
	}

	columns := []table.Column{
		{Title: "Name", Width: 15},
		{Title: "Last Modified", Width: 20},
		{Title: "Total Calls", Width: 12},
		{Title: "Success Rate", Width: 12},
		{Title: "Avg Duration", Width: 15},
		{Title: "Last Executed", Width: 20},
	}

	var rows []table.Row
	for _, action := range actions {
		stats, _ := statsManager.GetStats(action.Name)
		row := createActionRow(action, stats)
		rows = append(rows, row)
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(len(rows)+1),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("#7D56F4")).
		BorderBottom(true).
		Bold(true)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("#7D56F4")).
		Bold(false)
	t.SetStyles(s)

	return t, nil
}

func listGoctions(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var goctions []string
	for _, entry := range entries {
		if entry.IsDir() {
			goctions = append(goctions, entry.Name())
		}
	}
	return goctions, nil
}

func listActions(dir string) ([]Action, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var actions []Action
	for _, entry := range entries {
		if entry.IsDir() {
			mainFile := filepath.Join(dir, entry.Name(), "main.go")
			info, err := os.Stat(mainFile)
			if err == nil {
				actions = append(actions, Action{
					Name:         entry.Name(),
					LastModified: info.ModTime(),
				})
			}
		}
	}
	return actions, nil
}

func createActionRow(action Action, stats *stats.GoctionStats) table.Row {
	var totalCalls, successRate, avgDuration, lastExecuted string

	if stats != nil {
		totalCalls = fmt.Sprintf("%d", stats.TotalCalls)
		if stats.TotalCalls > 0 {
			successRate = fmt.Sprintf("%.2f%%", float64(stats.SuccessfulCalls)/float64(stats.TotalCalls)*100)
			avgDuration = (stats.TotalDuration / time.Duration(stats.TotalCalls)).String()
		} else {
			successRate = "N/A"
			avgDuration = "N/A"
		}
		if !stats.LastExecuted.IsZero() {
			lastExecuted = stats.LastExecuted.Format("2006-01-02 15:04:05")
		} else {
			lastExecuted = "Never"
		}
	} else {
		totalCalls = "0"
		successRate = "N/A"
		avgDuration = "N/A"
		lastExecuted = "Never"
	}

	return table.Row{
		action.Name,
		action.LastModified.Format("2006-01-02 15:04:05"),
		totalCalls,
		successRate,
		avgDuration,
		lastExecuted,
	}
}

func printGoctionStats(name string, stats *stats.GoctionStats) {
	statStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF")).
		PaddingLeft(2)

	nameStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#7D56F4"))

	fmt.Printf("%s:\n", nameStyle.Render(name))
	fmt.Printf("%s %d\n", statStyle.Render("Total calls:"), stats.TotalCalls)
	fmt.Printf("%s %d\n", statStyle.Render("Successful calls:"), stats.SuccessfulCalls)

	var successRate float64
	if stats.TotalCalls > 0 {
		successRate = float64(stats.SuccessfulCalls) / float64(stats.TotalCalls) * 100
	}
	fmt.Printf("%s %.2f%%\n", statStyle.Render("Success rate:"), successRate)
	var avgDuration time.Duration
	if stats.TotalCalls > 0 {
		avgDuration = stats.TotalDuration / time.Duration(stats.TotalCalls)
	}
	fmt.Printf("%s %s\n", statStyle.Render("Average duration:"), avgDuration)

	if !stats.LastExecuted.IsZero() {
		fmt.Printf("%s %s\n", statStyle.Render("Last executed:"), stats.LastExecuted.Format("2006-01-02 15:04:05"))
	} else {
		fmt.Printf("%s Never executed\n", statStyle.Render("Last executed:"))
	}
}

func getRecentLogs(logFile string, numLines int) ([]string, error) {
	file, err := os.Open(logFile)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Get the file size
	stat, err := file.Stat()
	if err != nil {
		return nil, err
	}

	// Calculate the size of the buffer we'll need
	bufferSize := int(stat.Size())
	if bufferSize > 1024*1024 { // limit to 1MB
		bufferSize = 1024 * 1024
	}

	// Read the file from the end
	buffer := make([]byte, bufferSize)
	start := stat.Size() - int64(bufferSize)
	if start < 0 {
		start = 0
	}
	_, err = file.ReadAt(buffer, start)
	if err != nil && err != io.EOF {
		return nil, err
	}

	// Find the last numLines lines
	lines := strings.Split(string(buffer), "\n")
	if len(lines) > numLines {
		lines = lines[len(lines)-numLines:]
	}

	// Trim any empty lines at the start
	for len(lines) > 0 && lines[0] == "" {
		lines = lines[1:]
	}

	return lines, nil
}

// ConfigView displays the current configuration
func ConfigView(cfg *config.Config) error {
	fmt.Println("Current Goction Configuration:")
	fmt.Printf("API Token: %s\n", cfg.APIToken)
	fmt.Printf("Goctions Directory: %s\n", cfg.GoctionsDir)
	fmt.Printf("Server Port: %d\n", cfg.Port)
	fmt.Printf("Log File: %s\n", cfg.LogFile)
	fmt.Printf("Stats File: %s\n", cfg.StatsFile)
	return nil
}

// ConfigReset resets the configuration to default values
func ConfigReset(cfg *config.Config) error {
	if err := config.Reset(); err != nil {
		return fmt.Errorf("failed to reset configuration: %w", err)
	}
	fmt.Println("Configuration has been reset to default values.")
	return nil
}

// ShowLogs displays recent log entries
func ShowLogs(cfg *config.Config) error {
	logs, err := getRecentLogs(cfg.LogFile, 20) // Show last 20 log entries
	if err != nil {
		return fmt.Errorf("failed to read logs: %w", err)
	}

	fmt.Println("Recent logs:")
	for _, log := range logs {
		fmt.Println(log)
	}
	return nil
}

// SelfUpdate updates Goction to the latest version
func SelfUpdate() error {
	// This is a placeholder. Implement actual self-update logic here.
	fmt.Println("Self-update functionality is not yet implemented.")
	fmt.Println("Please check https://github.com/goction/goction.git for updates.")
	return nil
}

type Action struct {
	Name         string
	LastModified time.Time
}
