package main

import (
	"bufio"
	"flag"
	"fmt"

	"os"
	"os/exec"
	"regexp"
	"strings"
	"sync"

	core "github.com/digininja/GitHunter/gitrob"

	"github.com/logrusorgru/aurora"
	"github.com/sirupsen/logrus"
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

By Robin Wood - https://digi.ninja - robin@digi.ninja
`

var Usage = func() {
	fmt.Fprintf(CommandLine.Output(), Banner)

	fmt.Fprintf(CommandLine.Output(), fmt.Sprintf("\nUsage: %s [options]\n", os.Args[0]))
	fmt.Fprintf(CommandLine.Output(), fmt.Sprintf("\nOptions:\n"))

	CommandLine.PrintDefaults()
}

var mainLogger = logrus.New()
var SomethingFound = false
var Commits map[string]Commit
var outputDestination *os.File

func main() {
	gitDirPtr := CommandLine.String("gitdir", ".", "Directory containing the repository")
	patternsFilePtr := CommandLine.String("patterns", "patterns.json", "File containing patterns to match")
	dumpPtr := CommandLine.Bool("dump", false, "Dump the commit details")
	nocoloursPtr := CommandLine.Bool("nocolours", false, "Set this to disable coloured output")
	helpPtr := CommandLine.Bool("help", false, "Show usage information")
	doGrepPtr := CommandLine.Bool("grep", false, "Grep files for content")
	debugPtr := CommandLine.String("debugLevel", "", "Debug options, I = Info, D = Full Debug")
	outputToPtr := CommandLine.String("output", "-", "File to write output to, - for standard out")

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

	if *outputToPtr == "-" {
		outputDestination = os.Stdout
	} else {
		var err error
		outputDestination, err = os.Create(*outputToPtr)
		if err != nil {
			mainLogger.Fatalf("Error creating the output file: %s", err)
		}
	}

	defer outputDestination.Close()

	fmt.Println(Banner)
	if *outputToPtr != "-" {
		fmt.Printf("Writing output to: %s\n", *outputToPtr)
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

	scanner := bufio.NewScanner(strings.NewReader(outputStr))
	first := true
	commit := Commit{}
	comment := ""

	Commits = make(map[string]Commit)

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
				Commits[commit.id] = commit
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
	Commits[commit.id] = commit
	var grepOutputRegexp *regexp.Regexp

	if *dumpPtr {
		pos := len(Commits)
		for _, c := range Commits {
			outputDestination.WriteString(fmt.Sprintf("Commit Number: %d\n", pos))
			c.PrintCommit()
			pos = pos - 1
		}
	} else {
		// Have to define this outside the next block so it is available later
		var revisionSliceChunks [][]string

		// Only need to pull the list of revisions out once
		// and only if doing a grep
		if doGrep {
			revList := ""
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

			// If there is a new line on the end, it creates an empty element at the end of the slice.
			// That is then passed as an empty argument to git which causes it to fail, even though it is nothing
			// So remove the trailing new line before splitting it and everything works.
			// Nearly an hour of debugging time to find this!
			if revList[len(revList)-1:] == "\n" {
				revList = revList[:len(revList)-1]
			}
			var revisionsSlice []string
			revisionsSlice = strings.Split(revList, "\n")
			// The higher this number, the more revisions grep will search at once
			// but the longer it will take doing it and so the output will look
			// jerky.
			chunkSize := 100
			for len(revisionsSlice) > chunkSize {
				revisionSliceChunks = append(revisionSliceChunks, revisionsSlice[0:chunkSize])
				revisionsSlice = revisionsSlice[chunkSize:len(revisionsSlice)]
			}
			revisionSliceChunks = append(revisionSliceChunks, revisionsSlice)

			// Naming these but not using the names at the moment. For more info see:
			// https://github.com/StefanSchroeder/Golang-Regex-Tutorial/blob/master/01-chapter2.markdown#named-matches
			grepOutputRegexpStr := "^(?P<ID>[a-f0-9]*):(?P<File>[^:]*):(?P<Message>.*)$"
			grepOutputRegexp = regexp.MustCompile(grepOutputRegexpStr)
		}

		var wg sync.WaitGroup

		done := make(chan bool)
		go printHits(done)

		for _, commit := range Commits {
			for _, signature := range CommentSignatures {
				/*
				   These are not guaranteed to all finish if the app finishes first.
				   Need to move them to channels and waitgroups

				   https://golangbot.com/channels/
				   https://golangbot.com/buffered-channels-worker-pools/
				*/

				// Check the commit messages
				wg.Add(1)
				go CommitMessageSearch(&wg, commit, signature)

				// Now checking for file contents
				if doGrep {
					//wg.Add(1)
					/*
					   Deliberately not doing this in a thread as git grep opens a lot of file handles
					   and so break things if ran concurrently.
					*/
					for _, chunk := range revisionSliceChunks {
						GrepSearch(commit, signature, chunk, gitDir, grepOutputRegexp)
					}
				}
			}
			// Finally check filenames
			wg.Add(1)
			go FilenameSearch(&wg, commit)
		}
		wg.Wait()
		close(hitsChannel)
		<-done
		if !SomethingFound {
			outputDestination.WriteString(fmt.Sprintln("Sorry, no interesting information found"))
		}
	}
}

type Hit struct {
	/*
		Might be better to pass a structure back
		so the output can do some nice processing on it,
		but for now, a string to print will do.

		hitType   string
		commit    Commit
		signature core.Signature
	*/
	output string
}

var hitsChannel = make(chan Hit, 10)

func printHits(done chan bool) {
	for hit := range hitsChannel {
		outputDestination.WriteString(fmt.Sprintf(hit.output))
		//fmt.Printf("Hit from the hits channel: %s\n", hit.commit.id)
	}
	done <- true
}

func FilenameSearch(wg *sync.WaitGroup, commit Commit) {
	for _, signature := range core.Signatures {
		for _, file := range commit.matchFiles {
			if signature.Match(file) {
				output := ""
				output += fmt.Sprintln(au.Bold(au.Blue("File Match")))
				output += fmt.Sprintf("Description: %s\n", signature.Description())
				if signature.Comment() != "" {
					output += fmt.Sprintf("Comment: %s\n", signature.Comment())
				}
				output += fmt.Sprintf("Hit on file: %s\n", file.Path)
				output += commit.GetCommitString()

				mainLogger.Debugf("Adding FilenameSearch result with commit ID %s to channel", commit.id)
				hit := Hit{output}
				hitsChannel <- hit

				SomethingFound = true
			}
		}
	}
	wg.Done()
}

func CommitMessageSearch(wg *sync.WaitGroup, commit Commit, signature CommentSignature) {
	if signature.Match(commit.comment) {
		output := ""
		output += fmt.Sprintln(au.Bold(au.Red("Commit Match")))
		output += fmt.Sprintf("Description: %s\n", signature.GetDescription())
		if signature.GetComment() != "" {
			output += fmt.Sprintf("Comment: %s\n", signature.GetComment())
		}
		output += commit.GetCommitString()

		mainLogger.Debugf("Adding CommitMessageSearch result with commit ID %s to channel", commit.id)
		hit := Hit{output}
		hitsChannel <- hit

		SomethingFound = true
	}

	wg.Done()
}

//func GrepSearch(wg *sync.WaitGroup, commit Commit, signature CommentSignature, revisionsSlice []string, gitDir string, grepOutputRegexp *regexp.Regexp) {
func GrepSearch(commit Commit, signature CommentSignature, revisionsSlice []string, gitDir string, grepOutputRegexp *regexp.Regexp) {
	var (
		cmdOut []byte
		err    error
	)

	cmdName := "git"
	cmdArgs := []string{}

	// Need to check for prefix of (?i) and if found, strip and add a -i to grep
	if signature.GetPattern()[0:4] == "(?i)" {
		pattern := strings.Replace(signature.GetPattern(), "(?i)", "", 1)
		cmdArgs = []string{"grep", "-i", "-E", pattern}
	} else {
		cmdArgs = []string{"grep", "-E", signature.GetPattern()}
	}

	for _, revisionId := range revisionsSlice {
		//mainLogger.Debugf("Adding revision to the command arguments: %s", revisionId)
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

	if err != nil {
		mainLogger.Debugf("err: %s", err.Error())
	}

	if err == nil {
		cmdOutStr := string(cmdOut)
		if cmdOutStr[len(cmdOutStr)-1:] == "\n" {
			cmdOutStr = cmdOutStr[:len(cmdOutStr)-1]
		}
		cmdOutMap := strings.Split(cmdOutStr, "\n")

		for _, commitLine := range cmdOutMap {
			output := ""

			output += fmt.Sprintln(au.Bold(au.Green("Grep Match")))
			//	mainLogger.Debugf("Commit line: %s", commitLine)
			//	mainLogger.Debugf("Commit line: %s", grepOutputRegexp)
			matchBits := grepOutputRegexp.FindStringSubmatch(commitLine)
			if len(matchBits) == 4 {
				commit := Commits[matchBits[1]]
				output += commit.GetCommitString()
				output += fmt.Sprintf("Match In File: %s\n", matchBits[2])
				output += fmt.Sprintf("Matching Line: %s\n\n", matchBits[3])

				mainLogger.Debugf("Adding GrepSearch result with commit ID %s to channel", commit.id)
				hit := Hit{output}
				hitsChannel <- hit
			}
		}

	} else if err.Error() == "exit status 1" {
		// Don't bail on 1
	} else {
		mainLogger.Fatal(fmt.Sprintf("There was an error running git grep command: %s", err))
	}
	//	outputStr := string(cmdOut)
	//	mainLogger.Debugf("Output from command: %s", outputStr)
	//wg.Done()
}
