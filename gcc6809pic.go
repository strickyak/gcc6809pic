package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

var NAME = flag.String("name", "noname", "module name to create")
var LIB1 = flag.String("lib1", "", "/home/strick/modoc/coco-shelf/gccretro/gcc/config/m6809/libgcc1.s")
var MEM = flag.Uint("mem", 256, "default memory size for module")

var CurrentArea string
var LibGcc1Defines []string
var Symbols = make(map[string]*Symbol)
var Areas = make(map[string]struct{})

type Symbol struct {
	Label string
	Area  string
}

type AsmLine struct {
	Label    string
	Opcode   string
	Args     string
	Remark   string
	Filename string
	LineNum  int
	Line     string
}

// Remark Lines:
// Blank Space, possibly followed by `;` or `*` and anything.
var RemarkLinePattern = regexp.MustCompile(`^\s*([;*].*)?$`)
var LabelLinePattern = regexp.MustCompile(`^([A-Za-z_.@$][A-Za-z0-9_.@$#?]*)\s*([*;].*)?$`)
var LabelColonLinePattern = regexp.MustCompile(`^\s*([A-Za-z_.@$][A-Za-z0-9_.@$#?]*)[:]\s*([*;].*)?$`)
var FullLinePattern = regexp.MustCompile(`^(([A-Za-z_.@$][A-Za-z0-9_.@$#?]*)[:]?)?\s+([A-Za-z0-9._=]+)\s*(.*?)\s*$`)

var NewFullLinePattern = regexp.MustCompile(`(\s*[A-Za-z_.@][A-Za-z0-9_.@$?]*[:]|^[A-Za-z_.@][A-Za-z0-9_.@$?]*)?(\s+[^ \t;]+)?(\s*.*)?\s*$`)

const WHITE_SET = "\t\n\v\f\r "

func SplitFullLine(s string) (label, op, arg string, ok bool) {
    if m := NewFullLinePattern.FindStringSubmatch(s); m != nil {
        label, _ = strings.CutSuffix(m[1], ":")
        op = strings.TrimLeft(m[2], WHITE_SET)
        arg = strings.TrimLeft(m[3], WHITE_SET)
        ok = true
        return
    }
    return
    // return "", "", Format(";;;NOT_NewFullLinePattern{%q}", s), false
}

func (o *AsmLine) String() string {
    return Format("%q: %s %v ;%q %s:%d", o.Label, o.Opcode, o.Args, o.Remark, o.Filename, o.LineNum)
}

func main() {
	log.SetFlags(0)
	flag.Parse()
	var z []*AsmLine

	for _, filename := range flag.Args() {
        PreSlurp(filename, false)
	}
    if *LIB1 != "" {
        PreSlurp(*LIB1, true)
    }

    for k, v := range LibProvides {
        log.Printf("NOTE %q provides %q", k, v)
    }
    for k, v := range LibRequires {
        log.Printf("NOTE %q requires %q", k, v)
    }

	for _, filename := range flag.Args() {
		z = append(z, Slurp(filename)...)
	}
	EmitPrelude()

	z2 := Tweak(z)
	Emit(z2)
	EmitPostlude()
}

/*
SymbolNotFound: "_udivhi3"
SymbolNotFound: "_umodhi3"
SymbolNotFound: "_mulhi3"
SymbolNotFound: "_divhi3"
SymbolNotFound: "_modhi3"
SymbolNotFound: "_ashrhi3"
SymbolNotFound: "_lshrhi3"
SymbolNotFound: "_ashlhi3"
*/

var ArithLib = regexp.MustCompile(`\b_(u?(div|mod|mul)|[al]sh[rl])hi3\b`)

func Slurp(filename string) (z []*AsmLine) {
	fd := Value(os.Open(filename))
	scanner := bufio.NewScanner(fd)
	i := 0
	for scanner.Scan() {
		i++
		line := scanner.Text()
		line = strings.TrimRight(line, " \t\r\n")

        if m := ArithLib.FindStringSubmatch(line); m != nil {
            LibGcc1Defines = append(LibGcc1Defines, m[0])
        }

		if m := RemarkLinePattern.FindStringSubmatch(line); m != nil {
			z = append(z, &AsmLine{
				Filename: filename,
				LineNum:  i,
				Line:     line,
				Remark:   line,
			})
			continue
		}
		if m := LabelLinePattern.FindStringSubmatch(line); m != nil {
			label := m[1]
			Symbols[label] = &Symbol{
				Label: label,
				Area:  CurrentArea,
			}
			z = append(z, &AsmLine{
				Filename: filename,
				LineNum:  i,
				Line:     line,
				Label:    label,
			})
			continue
		}
		if m := LabelColonLinePattern.FindStringSubmatch(line); m != nil {
			label := m[1]
			Symbols[label] = &Symbol{
				Label: label,
				Area:  CurrentArea,
			}
			z = append(z, &AsmLine{
				Filename: filename,
				LineNum:  i,
				Line:     line,
				Label:    label,
			})
			continue
		}
		if label, opcode, args, ok := SplitFullLine(line); ok {
		// if m := FullLinePattern.FindStringSubmatch(line); m != nil {
			// label, opcode, args := m[2], m[3], m[4]
            fmt.Printf(";nando; %q %q %q ====== %q\n", label, opcode, args, line)
			if opcode == ".area" {
				CurrentArea = FirstWord(args)
				Areas[CurrentArea] = struct{}{}
			}
			if label != "" {
				Symbols[label] = &Symbol{
					Label: label,
					Area:  CurrentArea,
				}
			}
			z = append(z, &AsmLine{
				Filename: filename,
				LineNum:  i,
				Line:     line,
				Label:    label,
				Opcode:   strings.ToLower(opcode),
				Args:     args,
			})
			continue
		}
		log.Fatalf("FATAL (%q:%d) gcc6809pic cannot parse %q", filename, i, line)
	}
	Check(scanner.Err(), filename)
	return
}

// var EmbeddedSymbolPattern = regexp.MustCompile(`^(.*)\b([_L][A-Za-z0-9_.]*)\b(.*)$`)
// var ExtendedSymbolArgPattern = regexp.MustCompile(`^([_L][A-Za-z0-9_.]*)(\s.*)?$`)
// var ImmediateSymbolArgPattern = regexp.MustCompile(`^([#][A-Za-z_][A-Za-z0-9_.]*)(\s.*)?$`)

var EmbeddedSymbolPattern = regexp.MustCompile(`^(.*)\b([_L][A-Za-z0-9_.]*)([-+]+[0-9]+)?(.*)$`)

var FindLabel= regexp.MustCompile(`^([_L][A-Za-z0-9_.]*)([-+]+[0-9]+)?(\s.*)?$`).FindStringSubmatch
var FindOctathorpe= regexp.MustCompile(`^[#]([A-Za-z_][A-Za-z0-9_.]*)([-+]+[0-9]+)?(\s.*)?$`).FindStringSubmatch

var FirstWordPattern = regexp.MustCompile(`^(\S+)(.*)$`)
var IdentifierPattern = regexp.MustCompile(`^([A-Za-z_][A-Za-z0-9_.]*)$`)

func FirstWord(s string) string {
	m := FirstWordPattern.FindStringSubmatch(s)
	if m == nil {
        return ""
		// log.Panicf("Expected a first argument: %q", s)
	}
	return m[1]
}

func CommentOut(a *AsmLine) *AsmLine {
	return &AsmLine{
		Remark: Format("*XXX* %s", a.Line),
	}
}

func TweakText(a *AsmLine, addend string, z []*AsmLine) []*AsmLine {
    add := func(op2 string, args2 string) {
			dup := *a
			dup.Label = ""
			dup.Opcode = op2
			dup.Args = args2
			z = append(z, &dup)
    }
    octothorpe := func() (string, uint, bool) {
		if m := FindOctathorpe(a.Args); m != nil {
            imm_sym, imm_addend := m[1], m[2]
            imm_addend_uint := AsciiToUintOrZero(imm_addend)
            return imm_sym, imm_addend_uint, true
        }
        return "UNUSED_212", 0, false
    }
	switch a.Opcode {

	case "jmp":
		dup := *a
		dup.Opcode = "lbra"
		dup.Args = Format("%s  ;PIC_FIXED: %s", dup.Args, a.Line)
		z = append(z, &dup)

	case "jsr":
		front := FirstWord(a.Args)
		if front == ",x" || front == ",y" || front == ",u" || front == ",s" {
			goto KEEP
		}

		isIdentifier := IdentifierPattern.FindStringSubmatch(front)
		if isIdentifier == nil {
			log.Panicf("(%s:%d) Case not handled: %q", a.Filename, a.LineNum, a.Line)
		}

		dup := *a // makes a copy of the struct
		dup.Opcode = "lbsr"
		dup.Args = Format("%s  ;PIC_FIXED: %s", front, a.Line)
		z = append(z, &dup)

    case "ldd":
        if x, y, ok := octothorpe(); ok {
            add("tfr", "x,d")
            add("leax", Format("%s+%d ;241;", x, y))
            add("exg", "d,x")
            //log.Panicf("TODO FindOctathorpe @@ %v" , a)
		} else if m := FindLabel(a.Args); m != nil {
            log.Panicf("TODO FindLabel @@ %v" , a)
		} else {
		    goto KEEP
        }

	case "lds", "std", "stx", "sty", "stu", "sts" :
		if m := FindOctathorpe(a.Args); m != nil {
            log.Panicf("TODO FindOctathorpe @@ %v" , a)
		} else if m := FindLabel(a.Args); m != nil {
            log.Panicf("TODO FindLabel @@ %v" , a)
		} else {
		    goto KEEP
        }

	case "ldx", "ldy", "ldu":
		if m := FindOctathorpe(a.Args); m != nil {
            imm_sym, imm_addend := m[1], m[2]
            imm_addend_uint := AsciiToUintOrZero(imm_addend)
			dup := *a
			dup.Opcode = Format("lea%c", a.Opcode[2])
			dup.Args = Format("%s+%d,pcr  ;PIC_IMM: %s", imm_sym, imm_addend_uint, a.Line)
			z = append(z, &dup)

		} else if m := FindLabel(a.Args); m != nil {
            imm_sym, imm_addend := m[1], m[2]
            imm_addend_uint := AsciiToUintOrZero(imm_addend)
			dup := *a
			dup.Opcode = Format("lea%c", a.Opcode[2])
			dup.Args = Format("%s+%d,pcr  ;PIC_LEA: %s", imm_sym, imm_addend_uint, a.Line )
			z = append(z, &dup)
			deref := *a
			deref.Label = ""
			deref.Opcode = a.Opcode
			deref.Args = Format(",%c  ;PIC_OP:", a.Opcode[2] )
			z = append(z, &deref)

		} else {
		    goto KEEP
        }
	default:
		goto KEEP
	}
	// return z with new stuff pushed onto it.
	return z

KEEP:
	// return z with the unchanged input pushed onto it.
	z = append(z, a)
	return z
}

var SymbolPlusMinus = regexp.MustCompile("^{}([+-])([0-9]+)$")

func LastChar(s string) byte {
    n := len(s)
    return s[n-1]
}

func TweakBss(a *AsmLine, symbol string, addend string, z []*AsmLine) []*AsmLine {
    add := func(format string, args ...any) {
        z = append(z, &AsmLine{Opcode: Format(format, args...)})
    }
    octothorpe := func() (string, uint, bool) {
		if m := FindOctathorpe(a.Args); m != nil {
            imm_sym, imm_addend := m[1], m[2]
            imm_addend_uint := AsciiToUintOrZero(imm_addend)
            return imm_sym, imm_addend_uint, true
        }
        return "UNUSED_212", 0, false
    }
    lpm := SymbolPlusMinus.FindStringSubmatch(front)

    if front == "{}" {
        add("%s  %s,y", a.Opcode, symbol)
    } else if front == "#{}" {
        switch a.Opcode {
        case "ldd":
            add("tfr  y,d")
            add("addd  #%s", a.Opcode, symbol)
        case "ldx", "ldy", "ldu", "lds":
            add("lea%c  %s,y", LastChar(a.Opcode), symbol)
        case "addd":
            add("pshs y")
            add("leay d,y")
            add("tfr y,d")
            add("puls y")
            z = append(z, a) // addd #_VAR

        default:
            panic(a.Line)
        }
    } else if lpm != nil {
        add("%s  %s%s%s,y", a.Opcode, symbol, lpm[1], lpm[2])
    } else {
        add("TODO_345 * %#v", a)
    }

	return z
}

func Tweak(a []*AsmLine) []*AsmLine {
    var z []*AsmLine
    for _, symbol := range LibGcc1Defines {
			Symbols[symbol] = &Symbol{
				Label: symbol,
				Area:  ".text",
			}
    }

    // Collect all the not-found symbols here.
    // If there are any of these at the end, panic.
    var symbolsNotFound []string

	for _, it := range a {
		switch it.Opcode {
		case ".module", ".area", ".globl":
			z = append(z, &AsmLine{
				Filename: it.Filename,
				LineNum:  it.LineNum,
				// Remark:   Format("%60s *** %s ***", "*", it.Line),
				Remark:   Format("%20s *** %s ***", "*", it.Line),
			})
		case ".ascii":
			z = append(z, it)
		case "":
			z = append(z, it)
		default:
			front := FirstWord(it.Args)
			m := EmbeddedSymbolPattern.FindStringSubmatch(front)
			if m != nil {
				symbol, addend := m[2], m[3]
				rec, ok := Symbols[symbol]
				if !ok {
                    symbolsNotFound = append(symbolsNotFound, symbol)
                    z = append(z, &AsmLine{Opcode:Format("Symbol__Not__Found__%q", symbol)})
                    continue
				}
				area := rec.Area
				front = strings.Replace(front, symbol, "{}", 1)
				switch area {
				case ".text", ".text.startup":
					z = TweakText(it, addend, z)
				case ".data":
					z = TweakText(it, addend, z)
				case ".bss":
					z = TweakBss(it, symbol, addend, z)
				default:
					log.Panicf("area not known: %q", area)
				}
			} else {
				z = append(z, it)
			}
		}
	}

	if len(symbolsNotFound) > 0 {
		for sym, rec := range Symbols {
						log.Printf("Symbol: %q = %v", sym, *rec)
		}
		for _, sym := range symbolsNotFound {
            log.Printf("SymbolNotFound::: %q", sym)
		}
        log.Panicf("%d symbols not found", len(symbolsNotFound))
    }
    return z
}

func Emit(a []*AsmLine) {
	for _, it := range a {
		switch {
		case it.Label == "" && it.Opcode == "" && it.Remark != "":
			// // fmt.Printf("%-60s *** %s ***\n", "", it.Remark)
			// fmt.Printf("%-60s %s\n", "", it.Remark)
			fmt.Printf("%s\n", it.Line)

		case it.Label == "" && it.Opcode == "" && it.Line == "":
			fmt.Printf("\n")

		case it.Label == "" && it.Opcode == "":
			// fmt.Printf("%-60s *** %s ***\n", "", it.Line)
			fmt.Printf("%s\n", it.Line)

		case it.Label != "":
			fmt.Printf("%-20s %-12s %s\n", it.Label+":", it.Opcode, it.Args)

		default:
			fmt.Printf("%-20s %-12s %s\n", "", it.Opcode, it.Args)
		}
	}
}

func EmitPrelude() {
    args := []string{"cpp"}
    for _, s := range LibGcc1Defines {
        args = append(args, Format("-D%s=1", s))
    }
    args = append(args, "/home/strick/modoc/coco-shelf/gccretro/gcc/config/m6809/libgcc1.s")
    cmd := &exec.Cmd{
        Path: "/usr/bin/cpp",
        Args: args,
        Stdin: nil,
        Stdout: os.Stdout,
        Stderr: os.Stderr,
    }
    e := cmd.Run()
    if e != nil {
        log.Fatalf("/usr/bin/cpp failed")
    }

	mem := Format("%d", *MEM)
	s := PRELUDE
	s = strings.Replace(s, "<-name->", *NAME, -1)
	s = strings.Replace(s, "<-mem->", mem, -1)
	fmt.Println(s)
}
func EmitPostlude() {
	fmt.Println(POSTLUDE)
}

const PRELUDE = `
*** Generated by gcc6809pic ***

 nam <-name->
 ttl <-name->

 ifp1
 use defsfile
 endc

tylg                set       Prgrm+Objct
atrv                set       ReEnt+rev
rev                 set       $01
edition             set       1


size                equ       <-mem->

                    mod       eom,name,tylg,atrv,start,size

name                fcs       /<-name->/
                    fcb       edition

start:              
                    clra
                    clrb
                    tfr d,x
                    tfr d,u
                    tfr dp,a
                    tfr d,y
                    tst ,y
                    lbsr _main
                    tfr x,d     ; X has return status, but we want it in B.
                    os9 F$Exit

undead:             bra undead  ; just get stuck because F$Exit should never have returned.
`

const POSTLUDE = `
    emod
eom equ *
    end
*** Generated by gcc6809pic ***
`
