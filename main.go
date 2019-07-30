package main

/*
This is a good reference for running commands:

https://nathanleclaire.com/blog/2014/12/29/shelled-out-commands-in-golang/
*/

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

import log "github.com/sirupsen/logrus"

type Commit struct {
	id        string
	author    string
	commit_by string
	comment   string
	files     []string
}

func main() {
	log.SetLevel(log.DebugLevel)
	gitDirPtr := flag.String("gitdir", "", "Directory containing the repository.")
	flag.Parse()

	gitDir := ""
	if *gitDirPtr != "" {
		gitDir = fmt.Sprintf("%s/.git", *gitDirPtr)
		log.Debugf("Checking to see if Git directory exists: %s", gitDir)
		if _, err := os.Stat(gitDir); os.IsNotExist(err) {
			log.Fatalf("The specified directory does not exist or is not a Git repository")
		}
	}

	var (
		cmdOut []byte
		err    error
	)
	cmdName := "git"
	cmdArgs := []string{"log", "--pretty=full", "--name-only", "--all"}
	if gitDir != "" {
		cmdArgs = append([]string{fmt.Sprintf("--git-dir=%s", gitDir)}, cmdArgs...)
	}

	log.Debugf("command arguments are: %u", cmdArgs)

	if cmdOut, err = exec.Command(cmdName, cmdArgs...).Output(); err != nil {
		log.Fatal(fmt.Sprintf("There was an error running git command: %s", err))
	}
	outputStr := string(cmdOut)
	// fmt.Println("Output from command: %s", outputStr)

	var commits []Commit

	scanner := bufio.NewScanner(strings.NewReader(outputStr))
	first := true
	commit := Commit{}
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "commit") {
			if first {
				first = false
			} else {
				commits = append(commits, commit)
				commit = Commit{}
			}
			log.Debugf("Commit ID: %s\n", line)
			commit.id = line
		} else if strings.HasPrefix(line, "Author") {
			commit.author = line
		} else if strings.HasPrefix(line, "Commit") {
			commit.commit_by = line
		} else if strings.HasPrefix(line, "Commit") {
			commit.commit_by = line
		} else {
			log.Debugf("Other: %s", scanner.Text())
		}
	}
	commits = append(commits, commit)

	for _, c := range commits {
		fmt.Println(c.id)
	}
}
