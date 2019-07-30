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
	dumpPtr := flag.Bool("dump", false, "Dump the commit details.")
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

	log.Debugf("Command arguments are: %s", cmdArgs)

	if cmdOut, err = exec.Command(cmdName, cmdArgs...).Output(); err != nil {
		log.Fatal(fmt.Sprintf("There was an error running git command: %s", err))
	}
	outputStr := string(cmdOut)
	// fmt.Println("Output from command: %s", outputStr)

	var commits []Commit

	scanner := bufio.NewScanner(strings.NewReader(outputStr))
	first := true
	commit := Commit{}
	comment := ""
	var files []string
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "commit") {
			if first {
				first = false
			} else {
				commit.comment = strings.TrimSpace(comment)
				commit.files = files
				commits = append(commits, commit)
				commit = Commit{}
				comment = ""
				files = nil
			}
			//	log.Debugf("Commit ID: %s\n", line)
			commit.id = line
		} else if strings.HasPrefix(line, "Author") {
			commit.author = line
		} else if strings.HasPrefix(line, "Commit") {
			commit.commit_by = line
		} else if strings.HasPrefix(line, "    ") {
			comment = comment + strings.TrimSpace(line) + "\n"
		} else {
			if line != "" {
				files = append(files, line)
			}
		}
	}
	commit.comment = strings.TrimSpace(comment)
	commit.files = files
	commits = append(commits, commit)

	if *dumpPtr {
		for i, c := range commits {
			fmt.Printf("Commit Number: %d\n", len(commits)-i)

			fmt.Printf("Commit ID: %s\n", c.id)
			fmt.Printf("Author: %s\n", c.author)
			fmt.Printf("Commit By: %s\n", c.commit_by)
			fmt.Printf("Comments: %s\n", c.comment)
			fmt.Println("Files:")
			for _, f := range c.files {
				fmt.Printf("  * %s\n", f)
			}
			fmt.Println()
		}
	}
}
