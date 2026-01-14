package main

// Slurp a config file with [] headers

import (
	"os"
	"regexp"
	"strings"
)

var FindBlankLine = regexp.MustCompile(`^\s*$`).FindStringSubmatch
var FindCommentLine = regexp.MustCompile(`^\s*[#]`).FindStringSubmatch
var FindHeaderLine = regexp.MustCompile(`^\s*[[](.*)[]]\s*$`).FindStringSubmatch

func ReadConfigFile(filename string) (z map[string][]string) {
    var header string

    s := string(Value(os.ReadFile(filename)))
    s = strings.ReplaceAll(s, "\r", "\n")
    lines := strings.Split(s, "\n")
    for _, line := range lines {
        if FindBlankLine(line) != nil {
            continue
        }
        if FindCommentLine(line) != nil {
            continue
        }
        if hl := FindHeaderLine(line); hl != nil {
            header = hl[1]
            continue
        }
        z[header] = append(z[header], line)
    }
    return
}
