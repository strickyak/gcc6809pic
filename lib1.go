package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
)

var LibDefines = make(map[string]string)
var LibProvides = make(map[string][]string)
var LibRequires = make(map[string][]string)
var LibContents = make(map[string][]string)

var MatchDefine = regexp.MustCompile(`#define\s+(\S+)\s+(.*)`).FindStringSubmatch
var MatchIfdef = regexp.MustCompile(`#ifdef\s+(\S+)`).FindStringSubmatch
var MatchEndif = regexp.MustCompile(`#endif`).FindStringSubmatch

var MatchEmbeddedIdentifier = regexp.MustCompile(`(.*)\b([_][A-Za-z0-9_.]+)\b(.*)?$`).FindStringSubmatch

func PreSlurp(filename string, ifdefMode bool) {
	fd := Value(os.Open(filename))
	scanner := bufio.NewScanner(fd)
	i := 0
	var w io.WriteCloser

	for scanner.Scan() {
		i++
		line := scanner.Text()
		line = strings.TrimRight(line, " \t\r\n")

		// Re-Iterate after success, if ever nested.
		for k, v := range LibDefines {
			line = strings.Replace(line, k, v, 1)
		}

		LibContents[filename] = append(LibContents[filename], line)
		if meip := MatchEmbeddedIdentifier(line); meip != nil {
			if meip[1] == "" {
				LibProvides[meip[2]] = append(
					LibProvides[meip[2]], filename)
			} else {
				LibRequires[meip[2]] = append(
					LibRequires[meip[2]], filename)
			}
		}

		md := MatchDefine(line)
		mi := MatchIfdef(line)
		me := MatchEndif(line)

		switch {
		case md != nil:
			LibDefines[md[1]] = md[2]
		case mi != nil:
			if w != nil {
				w.Close()
				w = nil
			}
			w = Value(os.Create(mi[1]))
		case me != nil:
			w.Close()
			w = nil
		default:
			if w != nil {
				fmt.Println(line)
			}
		}
	}
	Check(scanner.Err(), filename)
}
