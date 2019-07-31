package main

/*
This is a good reference for running commands:

https://nathanleclaire.com/blog/2014/12/29/shelled-out-commands-in-golang/
*/

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/michenriksen/gitrob/core"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

import log "github.com/sirupsen/logrus"

type Commit struct {
	id         string
	author     string
	authorDate time.Time
	commit     string
	commitDate time.Time
	comment    string
	matchFiles []core.MatchFile
}

func (c *Commit) PrintCommit() {
	fmt.Printf("Commit ID: %s\n", c.id)
	fmt.Printf("Author: %s\n", c.author)
	fmt.Printf("Author Date: %s\n", c.authorDate.String())
	fmt.Printf("Commit: %s\n", c.commit)
	fmt.Printf("Commit Date: %s\n", c.commitDate.String())
	fmt.Printf("Comments: %s\n", c.comment)
	fmt.Println("Files:")
	for _, f := range c.matchFiles {
		fmt.Printf("  * %s\n", f.Path)
	}
	fmt.Println()
}

func (c *Commit) AuthorDate(line string) {
	t, err := time.Parse("Mon Jan 2 15:04:05 2006 -0700", strings.TrimPrefix(line, "AuthorDate: "))
	if err == nil {
		c.authorDate = t
	} else {
		log.Debugf("Error parsing author date from: %s, err: %s\n", line, err)
	}
}

func (c *Commit) CommitDate(line string) {
	t, err := time.Parse("Mon Jan 2 15:04:05 2006 -0700", strings.TrimPrefix(line, "CommitDate: "))
	if err == nil {
		c.commitDate = t
	} else {
		log.Debugf("Error parsing commit date from: %s, err: %s\n", line, err)
	}
}

func main() {
	//log.SetLevel(log.DebugLevel)
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
	cmdArgs := []string{"log", "--pretty=fuller", "--name-only", "--all"}
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
	var matchFiles []core.MatchFile
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "commit") {
			if first {
				first = false
			} else {
				commit.comment = strings.TrimSpace(comment)
				commit.matchFiles = matchFiles
				matchFiles = nil
				commits = append(commits, commit)
				commit = Commit{}
				comment = ""
			}
			//	log.Debugf("Commit ID: %s\n", line)
			commit.id = strings.TrimPrefix(line, "commit ")
		} else if strings.HasPrefix(line, "Author:    ") {
			commit.author = strings.TrimPrefix(line, "Author:     ")
		} else if strings.HasPrefix(line, "AuthorDate:") {
			commit.AuthorDate(line)
		} else if strings.HasPrefix(line, "Commit:    ") {
			commit.commit = strings.TrimPrefix(line, "Commit:     ")
		} else if strings.HasPrefix(line, "CommitDate:") {
			commit.CommitDate(line)
		} else if strings.HasPrefix(line, "    ") {
			comment = comment + strings.TrimSpace(line) + "\n"
		} else {
			if line != "" {
				matchFile := core.NewMatchFile(line)
				matchFiles = append(matchFiles, matchFile)
			}
		}
	}

	commit.matchFiles = matchFiles
	commit.comment = strings.TrimSpace(comment)
	commits = append(commits, commit)

	commentsToWatch := []string{"mistake", "oops", "certificate", "keys"}

	commentsToWatchRegex := []*regexp.Regexp{}
	for _, comment := range commentsToWatch {
		var regexp = regexp.MustCompile(comment)
		commentsToWatchRegex = append(commentsToWatchRegex, regexp)
	}

	if *dumpPtr {
		for i, c := range commits {
			fmt.Printf("Commit Number: %d\n", len(commits)-i)
			c.PrintCommit()
		}
	} else {
		for _, c := range commits {
			for _, r := range commentsToWatchRegex {
				if r.FindStringIndex(c.comment) != nil {
					fmt.Printf("Commit Match\n")
					c.PrintCommit()
					fmt.Println()
				}
			}
		}
		for _, s := range core.Signatures {
			for _, c := range commits {
				for _, f := range c.matchFiles {
					if s.Match(f) {
						fmt.Printf("File Match\n")
						c.PrintCommit()
						fmt.Println()
					}
				}
			}
		}
	}
}
