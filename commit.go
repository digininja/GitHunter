package main

import (
	"fmt"
	"github.com/michenriksen/gitrob/core"
	"strings"
	"time"
)

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
	fmt.Printf(c.GetCommitString())
}

func (c *Commit) GetCommitString() string {
	output := ""

	output += fmt.Sprintf("Commit ID: %s\n", c.id)
	output += fmt.Sprintf("Author: %s\n", c.author)
	output += fmt.Sprintf("Author Date: %s\n", c.authorDate.String())
	output += fmt.Sprintf("Commit: %s\n", c.commit)
	output += fmt.Sprintf("Commit Date: %s\n", c.commitDate.String())
	output += fmt.Sprintf("Comments: %s\n", c.comment)
	output += fmt.Sprintln("Files:")
	for _, f := range c.matchFiles {
		output += fmt.Sprintf("  * %s\n", f.Path)
	}
	output += fmt.Sprintln()

	return output
}

func (c *Commit) AuthorDate(line string) {
	t, err := time.Parse("Mon Jan 2 15:04:05 2006 -0700", strings.TrimPrefix(line, "AuthorDate: "))
	if err == nil {
		c.authorDate = t
	} else {
		mainLogger.Debugf("Error parsing author date from: %s, err: %s\n", line, err)
	}
}

func (c *Commit) CommitDate(line string) {
	t, err := time.Parse("Mon Jan 2 15:04:05 2006 -0700", strings.TrimPrefix(line, "CommitDate: "))
	if err == nil {
		c.commitDate = t
	} else {
		mainLogger.Debugf("Error parsing commit date from: %s, err: %s\n", line, err)
	}
}
