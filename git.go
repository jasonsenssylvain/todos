package todos

import (
	"os/exec"
	"strings"
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

// func GitPrecommitHook(dir string) {
// 	dir = path.Join(dir, ".git/hooks/pre-push")
// 	bash := "#!/bin/bash"

// }
