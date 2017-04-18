package todos

import (
	"fmt"
	"os/exec"
	"path"
	"strings"
)

const (
	TODOS_DIRECTORY        = ".todos/"
	CLOSED_ISSUES_FILENAME = "closed.tx"
)

func GitDirectoryRoot() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	res, err := cmd.Output()
	if err != nil {
		return "", nil
	}
	return string(res), nil
}

func GitRemoteUrl() (string, error) {
	cmd := exec.Command("git", "ls-remote", "--get-url")
	res, err := cmd.Output()
	if err != nil {
		return "", err
	}
	remoteUrl := "https://github.com/" + string(res) + ".git\n"
	return remoteUrl, nil
}

func GitOwner() (string, error) {
	cmd := exec.Command("git", "ls-remote", "--get-url")
	res, err := cmd.Output()
	if err != nil {
		return "", err
	}
	remoteUrl := string(res)
	arr := strings.Split(remoteUrl, ":")
	arr = strings.Split(arr[1], "/")
	return arr[0], nil
}

func GitAdd(add string) error {
	cmd := exec.Command("git", "add", add)
	_, err := cmd.Output()
	return err
}

func GitBranch() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	res, err := cmd.Output()
	arr := strings.Split(string(res), "\n")
	return arr[0], err
}

func GitPrecommitHook(dir string) {
	dir = path.Join(dir, ".git/hooks/pre-push")
	bash := "#!/bin/bash"
	script := "git diff --name-only origin/master..HEAD | todos work"

	lines, _ := ReadLines(dir)
	if len(lines) == 0 {
		lines = append(lines, bash)
	}

	exists := false
	for _, line := range lines {
		if line == script {
			exists = true
			break
		}
	}

	if !exists {
		lines = append(lines, script)
	}
}

func GitCommitMessageHook(dir string) {
	f := path.Join(dir, TODOS_DIRECTORY, CLOSED_ISSUES_FILENAME)
	script := fmt.Sprintf("cat %s >> \"$1\"; rm -f %s", f, f)

	dir = path.Join(dir, ".git/hooks/commit-msg")
	bash := "#!/bin/bash"

	lines, _ := ReadLines(dir)
	if len(lines) == 0 {
		lines = append(lines, bash)
	}

	exists := false
	for _, line := range lines {
		if line == script {
			exists = true
			break
		}
	}

	if !exists {
		lines = append(lines, script)
	}
}
