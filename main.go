package main

/*
This is a good reference for running commands:

https://nathanleclaire.com/blog/2014/12/29/shelled-out-commands-in-golang/
*/

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/logrusorgru/aurora"
	"github.com/michenriksen/gitrob/core"
	"os"
	"os/exec"
	"strings"
)

import log "github.com/sirupsen/logrus"

var au aurora.Aurora

func main() {
	//log.SetLevel(log.DebugLevel)
	gitDirPtr := flag.String("gitdir", "", "Directory containing the repository.")
	dumpPtr := flag.Bool("dump", false, "Dump the commit details.")
	nocolours := flag.Bool("nocolours", false, "Set this to disable coloured output")

	flag.Parse()

	gitDir := ""
	if *gitDirPtr != "" {
		gitDir = fmt.Sprintf("%s/.git", *gitDirPtr)
		log.Debugf("Checking to see if Git directory exists: %s", gitDir)
		if _, err := os.Stat(gitDir); os.IsNotExist(err) {
			log.Fatalf("The specified directory does not exist or is not a Git repository")
		}
	}

	au = aurora.NewAurora(!*nocolours)

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

	if *dumpPtr {
		for i, c := range commits {
			fmt.Printf("Commit Number: %d\n", len(commits)-i)
			c.PrintCommit()
		}
	} else {
		for _, c := range commits {
			for _, s := range CommentSignatures {
				if s.Match(c.comment) {
					fmt.Println(au.Bold(au.Red("Commit Match")))
					fmt.Printf("Description: %s\n", s.Description())
					if s.Comment() != "" {
						fmt.Printf("Comment: %s\n", s.Comment())
					}
					c.PrintCommit()
					fmt.Println()
				}
			}

			for _, s := range core.Signatures {
				for _, f := range c.matchFiles {
					if s.Match(f) {
						fmt.Println(au.Bold(au.Red("File Match")))
						fmt.Printf("Description: %s\n", s.Description())
						if s.Comment() != "" {
							fmt.Printf("Comment: %s\n", s.Comment())
						}
						c.PrintCommit()
						fmt.Println()
					}
				}
			}
		}
	}
}
