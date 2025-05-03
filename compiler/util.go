package main

import (
	"regexp"
)

var pascalCaseRegex = regexp.MustCompile(`^([A-Z][a-z0-9]*)+$`)

var camelCaseRegex = regexp.MustCompile(`^[a-z]+(?:[A-Z][a-z0-9]*)*$`)

func IsPascalCase(s string) bool {
	return pascalCaseRegex.MatchString(s)
}

func IsCamelCase(s string) bool {
	return camelCaseRegex.MatchString(s)
}
