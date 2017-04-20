package main

import (
	"context"
	"crypto/sha1"
	"flag"
	"fmt"
	"log"
	"os/user"
	"regexp"
	"strings"

	"path"

	"time"

	"github.com/google/go-github/github"
	"github.com/skratchdot/open-golang/open"
	"golang.org/x/oauth2"
)

var tokenArg = flag.String("token", "", "Github token")
var resetArg = flag.Bool("reset", false, "reset github token")

func init() {
	user, err := user.Current()
	logOnError(err)
	HOME_DIRECTORY_CONFIG = user.HomeDir
	fmt.Println("home dir is " + HOME_DIRECTORY_CONFIG)
	flag.Parse()
}

func main() {
	root, err := GitDirectoryRoot()
	if err != nil {
		fmt.Println("must run todos inside a git repository")
	} else {
		if len(flag.Args()) < 1 {
			showHelp()
		} else {
			mode := flag.Args()[0]
			if mode == "setup" {
				setup(root)
			} else if mode == "work" {
				diff, err := ReadStdin()
				if len(diff) == 0 {
					diff, err = GitDiffFiles()
					logOnError(err)
				}
				fmt.Println(diff)
				diff = Map(func(s string) string { return path.Join(root, s) }, diff).([]string)
				fmt.Println(diff)
				work(root, diff)
			}
		}
	}
}

func showHelp() {
	fmt.Println("Unknown command.")
	fmt.Println("\t* setup: setup the current repository")
	fmt.Println("\t* work: runs todos and looks for todos in file")
}

func setup(root string) {
	config := OpenConfiguration(HOME_DIRECTORY_CONFIG)
	defer config.File.Close()

	if *tokenArg != "" {
		config.Config.GithubToken = *tokenArg
	} else if config.Config.GithubToken == "" || *resetArg {
		fmt.Println("Input Github access token:")
		open.Run(TOKEN_URL)
		var scanToken string
		fmt.Scanln(&scanToken)
		config.Config.GithubToken = scanToken
	}

	err := config.WriteConfiguration()
	if err != nil {
		fmt.Println(err)
	}

	localConfig := OpenConfiguration(root)
	defer localConfig.File.Close()

	if localConfig.Config.Owner == "" || localConfig.Config.Repo == "" || *resetArg {
		owner, _ := GitOwner()
		repo, _ := GitRemoteUrl()

		newOwner := ""
		newRepo := ""
		fmt.Printf("Enter the Github owner of the repo (Default: %s):\n", owner)
		fmt.Scanln(&newOwner)
		fmt.Printf("Enter the Github repo name (Default: %s):\n", repo)
		fmt.Scanln(&newRepo)

		if newOwner == "" || newOwner == "\n" {
			newOwner = owner
		}
		if newRepo == "" || newRepo == "\n" {
			newRepo = repo
		}

		localConfig.Config.Owner = newOwner
		localConfig.Config.Repo = newRepo
	}

	err = localConfig.WriteConfiguration()
	if err != nil {
		fmt.Println(err)
	}

	err = GitAdd(path.Join(root, TODOS_DIRECTORY))
	if err != nil {
		fmt.Println(err)
	}

	GitPrecommitHook(root)
	GitCommitMessageHook(root)
}

func work(root string, files []string) {
	config := OpenConfiguration(HOME_DIRECTORY_CONFIG)
	defer config.File.Close()

	localConfig := OpenConfiguration(root)
	defer localConfig.File.Close()

	if config.Config.GithubToken == "" {
		fmt.Println("Missing Github token. Set it in ~/.todos/conf.json")
	} else if localConfig.Config.Owner == "" || localConfig.Config.Repo == "" {
		fmt.Println("You need to setup the repo running 'todos setup'")
	} else {
		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: config.Config.GithubToken},
		)
		tc := oauth2.NewClient(oauth2.NoContext, ts)

		client := github.NewClient(tc)
		ctx := context.Background()

		owner := localConfig.Config.Owner
		repo := localConfig.Config.Repo

		existingRegex, err := regexp.Compile(ISSUE_URL_REGEX)
		logOnError(err)
		todoRegex, err := regexp.Compile(TODO_REGEX)
		logOnError(err)

		cacheFile := LoadIssueCache(root)
		cacheChanges := false

		closedIssues := []string{"", ""}

		for _, file := range files {
			relativeFilePath := pathDifference(root, file)

			fileIssuesCache := cacheFile.GetIssuesInFile(relativeFilePath)
			fileIssuesCacheCopy := fileIssuesCache

			removed := 0

			fmt.Println("check file: " + relativeFilePath)
			if relativeFilePath == root {
				continue
			}

			lines, err := ReadLines(file)
			logOnError(err)

			changes := false

			cb := make(chan Issue)
			issuesCount := 0

			for i, line := range lines {
				// fmt.Println(line)
				ex := existingRegex.FindString(line)
				todo := todoRegex.FindString(line)

				if ex != "" {
					for i, is := range fileIssuesCache {
						if is != nil && is.Hash == SHA1([]byte(ex)) {
							cacheChanges = true
							fileIssuesCacheCopy = fileIssuesCacheCopy.remove(i)
							removed++
						}
					}
				} else if todo != "" {
					issuesCount++
					go func(line int, cb chan Issue) {
						branch, _ := GitBranch()
						filename := path.Base(file)

						repos := strings.Split(repo, "/")
						repoName := repos[len(repos)-1]
						repoNames := strings.Split(repoName, ".")
						repoName = repoNames[0]

						body := fmt.Sprintf(ISSUE_BODY, filename, fmt.Sprintf(GITHUB_FILE_URL, owner, repoName, branch, relativeFilePath))
						fmt.Println("try to create issue")
						issue, _, err := client.Issues.Create(ctx, owner, repoName, &github.IssueRequest{Title: &todo, Body: &body})
						logOnError(err)

						if issue != nil {
							cb <- Issue{IssueUrl: *issue.HTMLURL, IssueNumber: *issue.Number, Line: line, File: relativeFilePath}
						}
					}(i, cb)
				}
			}

		loop:
			for issuesCount > 0 {
				select {
				case issue := <-cb:

					ref := fmt.Sprintf("[Issue: %s]", issue.IssueUrl)
					lines[issue.Line] = fmt.Sprintf("%s %s", lines[issue.Line], ref)
					fmt.Printf("[Todos] Created issue #%d\n", issue.IssueNumber)
					changes = true
					issuesCount--

					issue.Hash = SHA1([]byte(ref))
					cacheFile.Issues = append(cacheFile.Issues, &issue)
					cacheChanges = true
				case <-timeout(3 * time.Second):
					break loop
				}
			}

			closeCount := 0
			closeCb := make(chan Issue)
			for _, is := range fileIssuesCacheCopy {
				if is != nil {
					closeCount++
					go func(i Issue) {
						var closed string = "closed"
						repos := strings.Split(repo, "/")
						repoName := repos[len(repos)-1]
						repoNames := strings.Split(repoName, ".")
						repoName = repoNames[0]
						_, _, err := client.Issues.Edit(ctx, owner, repoName, is.IssueNumber, &github.IssueRequest{State: &closed})
						logOnError(err)
						closeCb <- i
					}(*is)
				}
			}

		loops:
			for closeCount > 0 {
				select {
				case is := <-closeCb:
					closeCount--
					issueClosing := fmt.Sprintf("[Todos] Closed issue #%d", is.IssueNumber)
					fmt.Println(issueClosing)
					closedIssues = append(closedIssues, issueClosing)
					cacheFile.RemoveIssue(is)
					cacheChanges = true
				case <-timeout(3 * time.Second):
					break loops
				}
			}

			if changes {
				logOnError(WriteFile(file, lines, false))
				GitAdd(file)
			} else {
				fmt.Println("[Todos] No new todos found")
			}
		}

		if cacheChanges {
			logOnError(cacheFile.WriteIssueCache())
			GitAdd(getIssuesCacheFIlePath(root))
		}
		if len(closedIssues) <= 2 {

			closedIssues = []string{}
		}

		logOnError(WriteFile(path.Join(root, TODOS_DIRECTORY, CLOSED_ISSUES_FILENAME), closedIssues, false))
	}
}

func pathDifference(p1, p2 string) string {

	return path.Join(strings.Split(p2, "/")[len(strings.Split(p1, "/")):]...)
}

func logOnError(err error) {

	if err != nil {
		log.Println("[Todos] Err:", err)
	}
}

func SHA1(data []byte) string {
	hash := sha1.New()
	hash.Write(data)
	return fmt.Sprintf("%x", hash.Sum(nil))
}

func timeout(i time.Duration) chan bool {

	t := make(chan bool)
	go func() {
		time.Sleep(i)
		t <- true
	}()

	return t
}
