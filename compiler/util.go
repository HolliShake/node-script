package main

import (
	"math"
	"regexp"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var pascalCaseRegex = regexp.MustCompile(`^[A-Z][a-z0-9]*(?:[A-Z][a-z0-9]*)*$`)

var camelCaseRegex = regexp.MustCompile(`^[a-z][a-z0-9]*(?:[A-Z][a-z0-9]*)*$`)

var snakeCaseRegex = regexp.MustCompile(`^[a-z][a-z0-9]*(?:_[a-z][a-z0-9]*)*$`)

var capitalCaseRegex = regexp.MustCompile(`^[A-Z][A-Z0-9]*$`)

var isFunctionCallRegex = regexp.MustCompile(`^[a-zA-Z0-9_]+\([^\)]*\)$`)

func IsPascalCase(s string) bool {
	return pascalCaseRegex.MatchString(s)
}

func IsCamelCase(s string) bool {
	return camelCaseRegex.MatchString(s)
}

func IsSnakeCase(s string) bool {
	return snakeCaseRegex.MatchString(s)
}

func IsCapitalCase(s string) bool {
	return capitalCaseRegex.MatchString(s)
}

func IsFunctionCall(s string) bool {
	return isFunctionCallRegex.MatchString(s)
}

func ToPascalCase(s string) string {
	return cases.Title(language.Und).String(s)
}

func ToSnakeCase(s string) string {
	// Convert PascalCase or camelCase to snake_case
	var result strings.Builder
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteRune('_')
		}
		result.WriteRune(r)
	}
	return strings.ToLower(result.String())
}

func JoinVariableName(origin, name string) string {
	if origin == "" {
		return name
	}
	return ToSnakeCase(origin) + "_" + name
}

func SizeOfInt(wholeNumber int64) byte {
	if wholeNumber >= math.MinInt8 && wholeNumber <= math.MaxInt8 {
		return 8
	}
	if wholeNumber >= math.MinInt16 && wholeNumber <= math.MaxInt16 {
		return 16
	}
	if wholeNumber >= math.MinInt32 && wholeNumber <= math.MaxInt32 {
		return 32
	}
	return 64
}
