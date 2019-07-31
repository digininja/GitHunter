package main

import (
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
	Description() string
	Comment() string
}

type SimpleCommentSignature struct {
	match       string
	description string
	comment     string
}

type PatternCommentSignature struct {
	match       *regexp.Regexp
	description string
	comment     string
}

func (s SimpleCommentSignature) Comment() string {
	return s.comment
}

func (s PatternCommentSignature) Comment() string {
	return s.comment
}

func (s SimpleCommentSignature) Description() string {
	return s.description
}

func (s PatternCommentSignature) Description() string {
	return s.description
}

func (s SimpleCommentSignature) Match(comment string) bool {
	return ContainsI(comment, s.match)
}

func (s PatternCommentSignature) Match(comment string) bool {
	return s.match.MatchString(comment)
}

var CommentSignatures = []CommentSignature{
	SimpleCommentSignature{
		match:       "keys",
		description: "Mention of keys",
		comment:     "",
	},
	SimpleCommentSignature{
		match:       "oops",
		description: "Mention of oops - could imply a mistake",
		comment:     "",
	},
	SimpleCommentSignature{
		match:       "mistake",
		description: "Mention of mistake",
		comment:     "",
	},
	PatternCommentSignature{
		// Prefix (?i) to make the regexp case insensitive
		match:       regexp.MustCompile("(?i)" + `[wx]whf`),
		description: "Regexp match",
		comment:     "",
	},
}
