package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"
)

// Repository represents a git repository configuration
type Repository struct {
	Name     string `json:"name"`
	Location string `json:"location"`
}

// RepoState represents the state of a repository
type RepoState struct {
	Name  string `json:"name"`
	State string `json:"state"`
}

// Config holds the application configuration
type Config struct {
	Repositories []Repository `json:"repositories"`
	PollInterval int          `json:"poll_interval"`
}

var (
	config     Config
	configLock sync.RWMutex
)

func loadConfig(configPath string) error {
	file, err := os.Open(configPath)
	if err != nil {
		return err
	}
	defer file.Close()

	var newConfig Config
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&newConfig); err != nil {
		return err
	}

	configLock.Lock()
	config = newConfig
	configLock.Unlock()
	return nil
}

func getRepoState(repo Repository) string {
	cmd := exec.Command("git", "status", "--porcelain=v1", "--branch")
	cmd.Dir = repo.Location
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

func pollRepositories() []RepoState {
	configLock.RLock()
	currentConfig := config
	configLock.RUnlock()

	states := []RepoState{}
	for _, repo := range currentConfig.Repositories {
		state := getRepoState(repo)
		states = append(states, RepoState{
			Name:  repo.Name,
			State: state,
		})
	}
	return states
}

func getOverallState(states []RepoState) string {
	for _, state := range states {
		if state.State == "dirty" {
			return "red"
		}
		if state.State == "ahead" {
			return "yellow"
		}
	}
	return "green"
}

func main() {
	userHome := os.Getenv("HOME")
	configFileName := filepath.Join(userHome, ".local", "share", "repowatcher", "config.json")
	configPath := configFileName

	if err := loadConfig(configPath); err != nil {
		fmt.Println("Error loading config file:", err)
		os.Exit(1)
	}

	// Handle SIGHUP for reloading configuration
	reloadChan := make(chan os.Signal, 1)
	signal.Notify(reloadChan, syscall.SIGHUP)
	go func() {
		for range reloadChan {
			if err := loadConfig(configPath); err != nil {
				fmt.Println("Error reloading config file:", err)
			} else {
				fmt.Println("Configuration reloaded.")
			}
		}
	}()

	for {
		states := pollRepositories()
		overallState := getOverallState(states)

		// Output for Waybar
		output := map[string]interface{}{
			"text":    overallState,
			"tooltip": states,
		}
		jsonOutput, _ := json.Marshal(output)
		fmt.Println(string(jsonOutput))

		configLock.RLock()
		interval := config.PollInterval
		configLock.RUnlock()
		time.Sleep(time.Duration(interval) * time.Second)
	}
}
