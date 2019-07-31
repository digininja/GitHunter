package main

import (
	"fmt"
	"github.com/michenriksen/gitrob/core"
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
