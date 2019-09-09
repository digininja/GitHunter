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
	"github.com/sirupsen/logrus"
	"os"
	"os/exec"
	"strings"
)

var au aurora.Aurora
var CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
var Banner = ` _______ __________________                           
(  ____ \\__   __/\__   __/                           
| (    \/   ) (      ) (                              
| |         | |      | |                              
| | ____    | |      | |                              
| | \_  )   | |      | |                              
| (___) |___) (___   | |                              
(_______)\_______/   )_(                              
                                                      
                   _       _________ _______  _______ 
|\     /||\     /|( (    /|\__   __/(  ____ \(  ____ )
| )   ( || )   ( ||  \  ( |   ) (   | (    \/| (    )|
| (___) || |   | ||   \ | |   | |   | (__    | (____)|
|  ___  || |   | || (\ \) |   | |   |  __)   |     __)
| (   ) || |   | || | \   |   | |   | (      | (\ (   
| )   ( || (___) || )  \  |   | |   | (____/\| ) \ \__
|/     \|(_______)|/    )_)   )_(   (_______/|/   \__/
`

var Usage = func() {
	fmt.Fprintf(CommandLine.Output(), Banner)

	fmt.Fprintf(CommandLine.Output(), fmt.Sprintf("\nUsage: %s [options]\n", os.Args[0]))
	fmt.Fprintf(CommandLine.Output(), fmt.Sprintf("\nOptions:\n"))

	CommandLine.PrintDefaults()
}

var mainLogger = logrus.New()

func main() {
	gitDirPtr := CommandLine.String("gitdir", ".", "Directory containing the repository")
	patternsFilePtr := CommandLine.String("patterns", "patterns.json", "File containing patterns to match")
	dumpPtr := CommandLine.Bool("dump", false, "Dump the commit details")
	nocoloursPtr := CommandLine.Bool("nocolours", false, "Set this to disable coloured output")
	helpPtr := CommandLine.Bool("help", false, "Show usage information")
	doGrepPtr := CommandLine.Bool("grep", false, "Grep files for content")
	debugPtr := CommandLine.String("debugLevel", "", "Debug options, I = Info, D = Full Debug")
	//testPtr := CommandLine.Bool("test", true, "Test stuff")

	CommandLine.Usage = Usage
	CommandLine.Parse(os.Args[1:])

	if *helpPtr {
		Usage()
		os.Exit(-1)
	}

	doGrep := *doGrepPtr

	switch strings.ToUpper(*debugPtr) {
	case "I":
		mainLogger.SetLevel(logrus.InfoLevel)
	case "D":
		mainLogger.SetLevel(logrus.DebugLevel)
	default:
		mainLogger.SetLevel(logrus.InfoLevel)
	}

	gitDir := ""
	if *gitDirPtr != "" {
		gitDir = *gitDirPtr
		if gitDir[len(gitDir)-1:] != "/" {
			gitDir = gitDir + "/"
		}
		gitDir = fmt.Sprintf("%s.git", gitDir)
		mainLogger.Debugf("Checking to see if Git directory exists: %s", gitDir)
		// A few options here, this is why I went with this one:
		// https://goruncode.com/how-to-check-if-a-file-exists/
		if _, err := os.Stat(gitDir); err != nil {
			mainLogger.Fatalf("The specified directory does not exist or does not contain a Git repository")
		}
	} else {
		Usage()
	}

	patternsFile := *patternsFilePtr
	mainLogger.Debugf("Checking to see if patterns file exists: %s", patternsFile)

	if _, err := os.Stat(patternsFile); err != nil {
		mainLogger.Fatalf("The specified patterns file does not exist")
	}

	ParsePatternsFile(patternsFile)

	au = aurora.NewAurora(!*nocoloursPtr)

	var (
		cmdOut []byte
		err    error
	)
	cmdName := "git"
	cmdArgs := []string{"log", "--pretty=fuller", "--name-only", "--all"}
	if gitDir != "" {
		cmdArgs = append([]string{fmt.Sprintf("--git-dir=%s", gitDir)}, cmdArgs...)
	}

	mainLogger.Debug("Getting all commit messages and files")
	mainLogger.Debugf("Command arguments are: %s", cmdArgs)

	if cmdOut, err = exec.Command(cmdName, cmdArgs...).Output(); err != nil {
		mainLogger.Fatal(fmt.Sprintf("There was an error running git command: %s", err))
	}
	outputStr := string(cmdOut)
	// mainLogger.Debugf("Output from command: %s", outputStr)

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
			//	mainLogger.Debugf("Commit ID: %s\n", line)
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
		somethingFound := false
		// Have to define this outside the next block so it is available later
		revList := ""

		// Only need to pull the list of revisions out once
		// and only if doing a grep

		if doGrep {
			var (
				revCmdOut []byte
				err       error
			)

			revCmdName := "git"
			revCmdArgs := []string{"rev-list", "--all"}
			if gitDir != "" {
				revCmdArgs = append([]string{fmt.Sprintf("--git-dir=%s", gitDir)}, revCmdArgs...)
			}

			mainLogger.Debug("Running git rev-list")
			mainLogger.Debugf("Command arguments are: %s", revCmdArgs)

			if revCmdOut, err = exec.Command(revCmdName, revCmdArgs...).Output(); err != nil {
				mainLogger.Fatal(fmt.Sprintf("There was an error running git rev-list command: %s", err))
			}
			revList = string(revCmdOut)
		}

		for _, c := range commits {
			for _, s := range CommentSignatures {

				// Check the commit messages

				if s.Match(c.comment) {
					fmt.Println(au.Bold(au.Red("Commit Match")))
					fmt.Printf("Description: %s\n", s.GetDescription())
					if s.GetComment() != "" {
						fmt.Printf("Comment: %s\n", s.GetComment())
					}
					c.PrintCommit()
					fmt.Println()
					somethingFound = true
				}

				// Now checking for file contents

				if doGrep {
					var (
						cmdOut []byte
						err    error
					)

					cmdName := "git"
					cmdArgs := []string{}
					//cmdName = "./echoit.sh"
					// need to check for prefix of (?i) and if found, strip and add a -i to grep
					mainLogger.Debugf("first 4 are: %s", s.GetPattern()[0:4])
					if s.GetPattern()[0:4] == "(?i)" {
						pattern := strings.Replace(s.GetPattern(), "(?i)", "", 1)
						cmdArgs = []string{"grep", "-i", "-E", pattern}
					} else {
						cmdArgs = []string{"grep", "-E", s.GetPattern()}
					}

					// If there is a new line on the end, it creates an empty element at the end of the slice.
					// That is then passed as an empty argument to git which causes it to fail, even though it is nothing
					// So remove the trailing new line before splitting it and everything works.
					// Nearly an hour of debugging time to find this!
					if revList[len(revList)-1:] == "\n" {
						revList = revList[:len(revList)-1]
					}
					revisionsMap := strings.Split(revList, "\n")

					for _, revisionId := range revisionsMap {
						fmt.Printf("adding: %s", revisionId)
						cmdArgs = append(cmdArgs, []string{revisionId}...)
					}

					if gitDir != "" {
						cmdArgs = append([]string{fmt.Sprintf("--git-dir=%s", gitDir)}, cmdArgs...)
					}

					mainLogger.Debug("Running a git grep")
					mainLogger.Debugf("Command arguments are: %s", cmdArgs)

					// If there are no matches, git will return 1
					// Matches have a return code of 0
					cmdOut, err = exec.Command(cmdName, cmdArgs...).Output()

					mainLogger.Debugf("cmdOut: %s", cmdOut)
					if err != nil {
						mainLogger.Debugf("err: %s", err.Error())
					}

					// Don't bail on 1
					if err == nil || err.Error() == "exit status 1" {
						fmt.Printf("ok")
					} else {
						mainLogger.Fatal(fmt.Sprintf("There was an error running git grep command: %s", err))
					}
					outputStr := string(cmdOut)
					mainLogger.Debugf("Output from command: %s", outputStr)
				}
				// git --git-dir=/home/robin/src/leakyrepo/.git grep -Ei "[vW]ulnerability" $(git --git-dir=/home/robin/src/leakyrepo/.git rev-list --all)

				// this needs to go in here
				// git grep -Ei "[vW]ulnerability" $(git rev-list --all)

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
						somethingFound = true
					}
				}
			}
		}
		if !somethingFound {
			fmt.Println("Sorry, no interesting information found")
		}
	}
}
