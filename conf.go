package main

// Slurp a config file with [] headers

import (
	"log"
	"os"
	"regexp"
	"strings"
)

var FindBlankLine = regexp.MustCompile(`^\s*$`).FindStringSubmatch
var FindCommentLine = regexp.MustCompile(`^\s*[#]`).FindStringSubmatch
var FindHeaderLine = regexp.MustCompile(`^\s*[[](.*)[]]\s*$`).FindStringSubmatch

type ConfigDefine struct {
	Key     string
	LineNum int
	Lines   []string
}

func ReadConfigFile(filename string) ( map[string]*ConfigDefine) {
    rules := make(map[string]*ConfigDefine)
	var cd *ConfigDefine

	s := string(Value(os.ReadFile(filename)))
	s = strings.ReplaceAll(s, "\r", "\n")
	lines := strings.Split(s, "\n")
	i := 0
	for _, line := range lines {
		i++
		if FindBlankLine(line) != nil {
			continue
		}
		if FindCommentLine(line) != nil {
			continue
		}
		if hl := FindHeaderLine(line); hl != nil {
            key := strings.TrimSpace(hl[1])
			cd = &ConfigDefine{
				Key:     key,
				LineNum: i,
			}
	        rules[key] = cd
            continue
		}
		if cd == nil {
			log.Panicf("Line %d not in any section: %q", i, line)
		}
		cd.Lines = append(cd.Lines, line)
	}
	return rules
}
