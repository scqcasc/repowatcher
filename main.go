package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

var configPath string
var onceMode bool

func init() {
	// Default config path
	defaultConfigPath := filepath.Join(os.Getenv("HOME"), ".local", "share", "repowatcher", "config.json")
	flag.StringVar(&configPath, "config", defaultConfigPath, "Path to configuration file")
	flag.BoolVar(&onceMode, "once", false, "Run a single check and exit")
	flag.Parse()
}

func loadConfig(filePath string) Config {
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		log.Fatalf("Failed to resolve absolute path for config file: %v", err)
	}

	file, err := os.Open(absPath)
	if err != nil {
		log.Fatalf("Failed to open config file: %v", err)
	}
	defer file.Close()

	var config Config
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		log.Fatalf("Failed to parse config file: %v", err)
	}

	// Ensure all repository paths are absolute
	for i := range config.Repositories {
		config.Repositories[i].Path = resolveAbsolutePath(config.Repositories[i].Path)
	}

	return config
}

func resolveAbsolutePath(repoPath string) string {
	absPath, err := filepath.Abs(repoPath)
	if err != nil {
		log.Fatalf("Failed to resolve absolute path for repository path %s: %v", repoPath, err)
	}
	return absPath
}

func getRepoState(repo Repository) string {
	cmd := exec.Command("git", "status", "--porcelain=v1", "--branch")
	cmd.Dir = repo.Path
	output, err := cmd.Output()
	if err != nil {
		return "error"
	}

	status := string(output)
	lines := strings.Split(status, "\n")

	// Check for branch ahead/behind state
	if len(lines) > 0 && strings.Contains(lines[0], "ahead") {
		return "ahead"
	}
	if len(lines) > 0 && strings.Contains(lines[0], "behind") {
		return "behind"
	}

	// Check for working tree cleanliness
	for _, line := range lines[1:] {
		if strings.TrimSpace(line) != "" {
			return "dirty" // If any line after the branch info contains data, the repo is dirty
		}
	}

	return "clean" // If no issues found, the repo is clean
}

func checkRepositories(repos []Repository) []RepoState {
	var wg sync.WaitGroup
	states := make([]RepoState, len(repos))

	for i, repo := range repos {
		wg.Add(1)
		go func(i int, repo Repository) {
			defer wg.Done()
			states[i] = RepoState{Name: repo.Name, State: getRepoState(repo)}
		}(i, repo)
	}
	wg.Wait()
	return states
}

func generateOutput(states []RepoState) {
	status := "green"
	var tooltips []string

	for _, state := range states {
		if state.State == "dirty" {
			status = "red"
		} else if state.State == "ahead" && status != "red" {
			status = "yellow"
		}
		tooltips = append(tooltips, fmt.Sprintf("%s: %s", state.Name, state.State))
	}

	output := struct {
		Text    string `json:"text"`
		Class   string `json:"class"`
		Tooltip string `json:"tooltip"` // Changed to string
	}{
		Text:    status,
		Class:   status,
		Tooltip: strings.Join(tooltips, "\n"), // Combine array into a single string
	}

	jsonOutput, err := json.Marshal(output)
	if err != nil {
		log.Fatalf("Failed to marshal JSON: %v", err)
	}

	fmt.Println(string(jsonOutput))
}

func checkRepositoriesAndOutput(config Config) {
	states := checkRepositories(config.Repositories)
	generateOutput(states) // Already prints valid JSON
}

func main() {
	// Load configuration
	config := loadConfig(configPath)

	if onceMode {
		// Run once and exit
		checkRepositoriesAndOutput(config)
		return
	}

	// Daemon mode
	for {
		checkRepositoriesAndOutput(config)
		time.Sleep(time.Duration(config.Interval) * time.Second)
	}
}
