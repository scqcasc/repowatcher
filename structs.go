package main

type Repository struct {
	Name string `json:"name"`
	Path string `json:"location"`
}

type Config struct {
	Repositories []Repository `json:"repositories"`
	Interval     int          `json:"interval"`
}

type RepoState struct {
	Name  string `json:"name"`
	State string `json:"state"`
}
