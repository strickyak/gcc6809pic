package main

import (
	"bufio"
	//"fmt"
	//"io"
	"log"
	"os"
	"regexp"
	"strings"
)

var LibProvidedBy = make(map[string]string)
var LibContents = make(map[string][]string)

var MatchIfdef = regexp.MustCompile(`#ifdef\s+(\S+)`).FindStringSubmatch
var MatchEndif = regexp.MustCompile(`#endif`).FindStringSubmatch

var MatchLabel = regexp.MustCompile(`^([A-Za-z0-9_.]+)[:]?`).FindStringSubmatch

var MatchEmbeddedIdentifier = regexp.MustCompile(`(.*)\b([_][A-Za-z0-9_.]+)\b(.*)?$`).FindStringSubmatch

func LoadLibrary() {
	log.Printf("LIBLINE HERE")
	tag := "none"
	fd := Value(os.Open(*LIB1))
	scanner := bufio.NewScanner(fd)
	i := 0

	for scanner.Scan() {
		i++
		line := scanner.Text()
		line = strings.TrimRight(line, " \t\r\n")

		mi := MatchIfdef(line)
		me := MatchEndif(line)
		ml := MatchLabel(line)
		log.Printf("LIBLINE %03d. %q mi %v me %v ml %v", i, line, mi, me, ml)
		switch {

		case mi != nil:
			tag = mi[1]

		case me != nil:
			tag = "none"

		case ml != nil:
			LibProvidedBy[ml[1]] = tag
			fallthrough

		default:
			LibContents[tag] = append(LibContents[tag], line)
		}
	}
	Check(scanner.Err(), *LIB1)
}
