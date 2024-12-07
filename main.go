package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
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

func getRepoState(repo Repository) string {
	cmd := exec.Command("git", "status", "--porcelain=v1", "--branch")
	cmd.Dir = repo.Location
	output, err := cmd.Output()
	if err != nil {
		return "error"
	}

	status := string(output)
	if strings.Contains(status, "ahead") {
		return "ahead"
	}
	if strings.Contains(status, "behind") {
		return "behind"
	}
	if strings.TrimSpace(status) == "" {
		return "clean"
	}
	return "dirty"
}

func pollRepositories(config Config) []RepoState {
	states := []RepoState{}
	for _, repo := range config.Repositories {
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
	configFile, err := os.Open("config.json")
	if err != nil {
		fmt.Println("Error opening config file:", err)
		os.Exit(1)
	}
	defer configFile.Close()

	var config Config
	decoder := json.NewDecoder(configFile)
	err = decoder.Decode(&config)
	if err != nil {
		fmt.Println("Error decoding config file:", err)
		os.Exit(1)
	}

	for {
		states := pollRepositories(config)
		overallState := getOverallState(states)

		// Output for Waybar
		output := map[string]interface{}{
			"text":    overallState,
			"tooltip": states,
		}
		jsonOutput, _ := json.Marshal(output)
		fmt.Println(string(jsonOutput))

		time.Sleep(time.Duration(config.PollInterval) * time.Second)
	}
}
