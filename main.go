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

type Repository struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

type Config struct {
	Repositories []Repository `json:"repositories"`
	Interval     int          `json:"interval"`
}

type RepoState struct {
	Name  string `json:"name"`
	State string `json:"state"`
}

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
	cmd := exec.Command("git", "-C", repo.Path, "status", "--porcelain", "-b")
	output, err := cmd.Output()
	if err != nil {
		log.Printf("Error checking repository %s: %v", repo.Name, err)
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

func generateOutput(states []RepoState) string {
	status := "green"
	for _, state := range states {
		if state.State == "dirty" {
			status = "red"
			break
		} else if state.State == "ahead" {
			status = "yellow"
		}
	}
	output := struct {
		Text    string      `json:"text"`
		Tooltip []RepoState `json:"tooltip"`
	}{
		Text:    status,
		Tooltip: states,
	}
	result, _ := json.Marshal(output)
	return string(result)
}

func checkRepositoriesAndOutput(config Config) {
	states := checkRepositories(config.Repositories)
	fmt.Println(generateOutput(states))
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
