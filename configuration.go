package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
)

const (
	TODOS_DIRECTORY        = ".todos/"
	HOME_DIRECTORY_CONFIG  = "my home dir"
	TODOS_CONF_FILENAME    = "conf.json"
	ISSUE_CACHE_FILENAME   = "issues.json"
	CLOSED_ISSUES_FILENAME = "closed.txt"
	TOKEN_URL              = "https://github.com/settings/tokens/new?scopes=repo,public_repo"
)

const (
	ISSUE_URL_REGEX = "\\[(Issue:[^\\]]*)\\]"
	TODO_REGEX      = "TODO(\\([^)]+\\))?:(.*)"
	ISSUE_BODY      = "On file: [%s](%s)"
	GITHUB_FILE_URL = "https://github.com/%s/%s/blob/%s/%s"
)

type Configuration struct {
	GithubToken string `json:"github_token,omitempty"`
	Owner       string `json:"github_owner,omitempty"`
	Repo        string `json:"github_repo,omitempty"`
}

type ConfFile struct {
	Config Configuration
	File   *os.File
}

type Issue struct {
	File        string
	Hash        string
	IssueNumber int
	Line        int
	IssueUrl    string
}

type Issues []*Issue

func (slice Issues) Len() int {
	return len(slice)
}

func (slice Issues) Swap(i, j int) {
	slice[i], slice[j] = slice[j], slice[i]
}

func (slice Issues) Less(i, j int) bool {
	return slice[i].Hash < slice[j].Hash
}

func (slice Issues) remove(i int) Issues {
	copy(slice[:i], slice[:i+1])
	slice[len(slice)-1] = nil
	return slice[:len(slice)-1]
}

type IssueCacheFile struct {
	Issues Issues
	File   *os.File
}

func OpenConfiguration(dir string) *ConfFile {
	dir = path.Join(dir, TODOS_DIRECTORY)

	err := os.MkdirAll(dir, 0777)
	if err != nil {
		fmt.Println(err)
	}

	f, err := os.OpenFile(path.Join(dir, TODOS_CONF_FILENAME), os.O_RDWR|os.O_CREATE, 0660)
	if err != nil {
		fmt.Println(err)
	}

	conf := Configuration{}
	json.NewDecoder(f).Decode(&conf)

	return &ConfFile{conf, f}
}

func (conf *ConfFile) WriteConfiguration() error {
	conf.File.Truncate(0)
	conf.File.Seek(0, 0)
	err := json.NewEncoder(conf.File).Encode(conf.Config)
	if err != nil {
		fmt.Println(err)
	}
	return conf.File.Close()
}

func getIssuesCacheFIlePath(dir string) string {
	return path.Join(dir, TODOS_DIRECTORY, ISSUE_CACHE_FILENAME)
}

func LoadIssueCache(dir string) *IssueCacheFile {
	filepath := getIssuesCacheFIlePath(dir)

	f, err := os.OpenFile(filepath, os.O_RDWR|os.O_CREATE, 0660)
	if err != nil {
		fmt.Println(err)
	}

	issues := Issues{}
	json.NewDecoder(f).Decode(&issues)

	return &IssueCacheFile{issues, f}
}

func (cache *IssueCacheFile) GetIssuesInFile(file string) Issues {
	array := Issues{}

	for _, is := range cache.Issues {
		if is != nil && is.File == file {
			array = append(array, is)
		}
	}
	return array
}

func (cache *IssueCacheFile) RemoveIssue(issue Issue) {
	for i, is := range cache.Issues {
		if is != nil && issue == *is {
			cache.Issues.remove(i)
		}
	}
}

func (cache *IssueCacheFile) WriteIssueCache() error {
	cache.File.Truncate(0)
	cache.File.Seek(0, 0)

	err := json.NewEncoder(cache.File).Encode(cache.Issues)
	if err != nil {
		fmt.Println(err)
	}

	return cache.File.Close()
}
