package main

import (
	"fmt"
	"math"
	"os"
	"runtime"
	"strings"
)

var PADDING = 3

func RaiseSystemError(message any) {
	_, fileName, line, _ := runtime.Caller(1)
	err := fmt.Sprintf("DEBUG(%s:%d) | [ERROR] %s", fileName, line, fmt.Sprint(message))
	fmt.Fprint(os.Stderr, err)
	// Collect and free the memory
	CollectAndFree()
	// Exit the program
	os.Exit(1)
}

func RaiseLanguageCompileError(file string, data []rune, message string, position TPosition) {
	_, fileName, line, _ := runtime.Caller(1)
	lines := strings.Split(string(data), "\n")
	start := int(math.Max((float64(position.SLine)-1)-float64(PADDING), 0))
	ended := int(math.Min((float64(position.ELine)+0)+float64(PADDING), float64(len(lines))))
	fmtMessage := fmt.Sprintf("DEBUG(%s:%d) | [ERROR] %s:%d:%d: %s\n", fileName, line, file, position.SLine, position.SColm, message)
	strEnded := fmt.Sprintf("%d", int(ended))
	for i := start; i < ended; i++ {
		strStart := fmt.Sprintf("%d", i+1)
		strDiffs := len(strEnded) - len(strStart)
		fmtMessage += fmt.Sprintf("%s%s | ", strings.Repeat(" ", strDiffs), strStart)

		if (i+1) >= position.SLine && (i+1) <= position.ELine {
			fmtMessage += " ~ "
		} else {
			fmtMessage += "   "
		}
		fmtMessage += lines[i]
		if (i + 1) <= ended {
			fmtMessage += "\n"
		}
	}
	fmt.Fprint(os.Stderr, fmtMessage)
	// Collect and free the memory
	CollectAndFree()
	// Exit the program
	os.Exit(1)
}
