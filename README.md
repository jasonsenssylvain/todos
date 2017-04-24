# todos

    a tool that can generate issues into your github project if where is a TODO comment in your code.

### how to use:
    go get github.com/jasoncodingnow/todos
    in your github project: todos setup or todos work
    
    for example:
    
    //TODO: need to fix that stupid issue
    func XXX() error {
      YYY...
    }

#### Day 1:

git.go - tool to interative with git

  * GitProjectRoot()
  * GitRemoteUrl()
  * GitOwner()
  * GitAdd()
  * GitBranch()
  * SetupGitPrecommitHook()
  * SetupGitCommitMsgHook()

#### Day 2:

io.go - tool to interative with local code file

  * ReadLines()
  * ReadStdin()
  * WriteFile()
