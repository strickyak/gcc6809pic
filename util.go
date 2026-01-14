package main

import (
	"fmt"
	"log"
	"runtime/debug"
    "strconv"
)

var Format = fmt.Sprintf
var Panicf = log.Panicf

func AsciiToUintOrZero(s string) uint {
    z, err := strconv.ParseUint(s, 10, 32)
    if err != nil {
        return 0
    }
    return uint(z)
}

func Value[T any](value T, err error) T {
	Check(err)
	return value
}

func Check(err error, args ...any) {
	if err != nil {
		s := fmt.Sprintf("Check Fails: %v", err)
		for _, x := range args {
			s += fmt.Sprintf(" ; %v", x)
		}
		s += "\n[[[[[[\n" + string(debug.Stack()) + "\n]]]]]]\n"
		log.Panic(s)
	}
}
