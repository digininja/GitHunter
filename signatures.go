package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
)

// Case insensitive contains function
func ContainsI(a string, b string) bool {
	return strings.Contains(
		strings.ToLower(a),
		strings.ToLower(b),
	)
}

type CommentSignature interface {
	Match(comment string) bool
	GetDescription() string
	GetComment() string
}

type SimpleCommentSignature struct {
	Pattern     string
	Description string
	Comment     string
}

func (p *PatternCommentSignature) CompileRegexp() {
	mainLogger.Debug("In here")
	p.Regexp = regexp.MustCompile(p.Pattern)
}

type PatternCommentSignature struct {
	Regexp      *regexp.Regexp
	Pattern     string
	Description string
	Comment     string
}

func (s SimpleCommentSignature) GetComment() string {
	return s.Comment
}

func (s PatternCommentSignature) GetComment() string {
	return s.Comment
}

func (s SimpleCommentSignature) GetDescription() string {
	return s.Description
}

func (s PatternCommentSignature) GetDescription() string {
	return s.Description
}

func (s SimpleCommentSignature) Match(comment string) bool {
	return ContainsI(comment, s.Pattern)
}

func (s PatternCommentSignature) Match(comment string) bool {
	return s.Regexp.MatchString(comment)
}

var CommentSignatures = []CommentSignature{}

type Patterns struct {
	Patterns []PatternCommentSignature
	Simples  []SimpleCommentSignature
}

func ParseConfig() bool {
	var configstruct Patterns

	jsonFile, err := os.Open("patterns.json")

	if err != nil {
		mainLogger.Fatalf("Error opening patterns file: %s", err)
	}

	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)

	json.Unmarshal(byteValue, &configstruct)

	mainLogger.Debugf("Patterns: %u)", configstruct)

	for _, pattern := range configstruct.Simples {
		CommentSignatures = append(CommentSignatures, pattern)
	}

	for _, pattern := range configstruct.Patterns {
		pattern.CompileRegexp()
		CommentSignatures = append(CommentSignatures, pattern)
	}

	mainLogger.Debugf("Merged: %u", CommentSignatures)

	for _, pattern := range CommentSignatures {
		mainLogger.Debugf("Description is: %s", pattern.GetDescription())
	}

	mainLogger.Debug("done, out of here")
	return true
}

/*
var CommentSignatures = []CommentSignature{
	SimpleCommentSignature{
		Pattern:     "keys",
		Description: "Mention of keys",
		Comment:     "",
	},
	SimpleCommentSignature{
		Pattern:     "oops",
		Description: "Mention of oops - could imply a mistake",
		Comment:     "",
	},
	SimpleCommentSignature{
		Pattern:     "mistake",
		Description: "Mention of mistake",
		Comment:     "",
	},
	PatternCommentSignature{
		// Prefix (?i) to make the regexp case insensitive
		Pattern:     regexp.MustCompile("(?i)" + `[wx]whf`),
		Description: "Regexp match",
		Comment:     "",
	},
}
*/
