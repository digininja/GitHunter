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
	GetPattern() string
}

type SimpleCommentSignature struct {
	Pattern     string
	Description string
	Comment     string
}

func (p *PatternCommentSignature) CompileRegexp() {
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

func (s SimpleCommentSignature) GetPattern() string {
	return s.Pattern
}

func (s PatternCommentSignature) GetPattern() string {
	return s.Pattern
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

type JSONPatterns struct {
	Patterns []PatternCommentSignature
	Simples  []SimpleCommentSignature
}

func ParsePatternsFile(patternsFile string) bool {
	mainLogger.Debug("Starting JSON patterns file parsing")
	var jsonPatterns JSONPatterns

	jsonFile, err := os.Open(patternsFile)

	if err != nil {
		mainLogger.Fatalf("Error opening patterns file: %s", err)
	}

	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)

	json.Unmarshal(byteValue, &jsonPatterns)

	for _, pattern := range jsonPatterns.Simples {
		CommentSignatures = append(CommentSignatures, pattern)
	}

	for _, pattern := range jsonPatterns.Patterns {
		// Doing this to compile the string in the JSON file into a
		// regexp that can then be used by the match function
		pattern.CompileRegexp()
		CommentSignatures = append(CommentSignatures, pattern)
	}

	mainLogger.Debug("JSON patterns file parsing complete")
	return true
}
