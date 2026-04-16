// Compile with gcc6809 with these:
//  -fwhole-program -fomit-frame-pointer -ffixed-y

package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	//"os/exec"
	"regexp"
	"runtime/debug"
	"strings"
)

var NAME = flag.String("name", "noname", "module name to create")
var LIB1 = flag.String("lib1", "/home/strick/modoc/coco-shelf/gcc6809pic/libgcc1.library", "libgcc1.s location")
var MEM = flag.Uint("mem", 256, "default memory size for module")
var RULES = flag.String("rules", "rules.txt", "text file containing rewrite rules")

var CurrentArea string
var LibNeeded = make(map[string]bool)
var ProviderEmitted = make(map[string]bool)
var Symbols = make(map[string]*Symbol)
var Areas = make(map[string]struct{})

type Symbol struct {
	Label string
	Area  string
}

type AsmLine struct {
	Label       string
	Opcode      string
	Args        string
	Remark      string
	Filename    string
	LineNum     int
	Line        string
	How         string
	Parameter   string // from first argument
	Kind        string // of first argument
	Index       int
	GenericArgs string
}

// Remark Lines:
// Blank Space, possibly followed by `;` or `*` and anything.
var RemarkLinePattern = regexp.MustCompile(`^\s*([;*].*)?$`)
var LabelLinePattern = regexp.MustCompile(`^([A-Za-z_.@$][A-Za-z0-9_.@$#?]*)\s*([*;].*)?$`)
var LabelColonLinePattern = regexp.MustCompile(`^\s*([A-Za-z_.@$][A-Za-z0-9_.@$#?]*)[:]\s*([*;].*)?$`)

var NewFullLinePattern = regexp.MustCompile(`^(\s*[A-Za-z_.@][A-Za-z0-9_.@$?]*[:]|[A-Za-z_.@][A-Za-z0-9_.@$?]*)?(\s+[^ ;]+)?(\s*[^ ;]+)?(.*)$`)

const WHITE_SET = "\t\n\v\f\r "

var SymbolInArgsPattern = regexp.MustCompile(`[A-Za-z_.@][A-Za-z0-9_.@$#?]+`)

func SplitFullLine(s string) (label, op, arg, remark string, ok bool) {
	if m := NewFullLinePattern.FindStringSubmatch(s); m != nil {
		label, _ = strings.CutSuffix(m[1], ":")
		op = strings.TrimLeft(m[2], WHITE_SET)
		arg = strings.TrimLeft(m[3], WHITE_SET)
		remark = strings.TrimLeft(m[4], WHITE_SET)
		ok = true
		return
	}
	return
}

func (o *AsmLine) String() string {
	return Format("%q: { %s | %v };%q -:- \\%s/", o.Label, o.Opcode, o.Args, o.Remark, o.How)
}

func DeclareLibNeeded(symbol string) {
	Symbols[symbol] = &Symbol{
		Label: symbol,
		Area:  ".text",
	}
	LibNeeded[symbol] = true
}

func main() {
	flag.Parse()

	logFilename := *NAME + ".pic.log"
	logWriter := Value(os.Create(logFilename))
	log.SetOutput(logWriter)
	log.SetFlags(0)

	defer func() {
		r := recover()
		if r != nil {
			debug.PrintStack()
			fmt.Fprintf(os.Stderr, "\nFATAL: %v\n", r)
			fmt.Fprintf(os.Stderr, "See %q for details.\n", logFilename)
			os.Exit(13)
		}
	}()

	rules := ReadConfigFile()
	log.Printf("LIBLINE HERE")
	if *LIB1 != "" {
		log.Printf("LIBLINE HERE")
		LoadLibrary()
	}

	for k, v := range LibProvidedBy {
		log.Printf("NOTE %q provided by %q", k, v)
	}

	var asms []*AsmLine
	for _, filename := range flag.Args() {
		asms = append(asms, Slurp(filename)...)
	}

	Examinem(asms)

	EmitPrelude()
	Changem(asms, rules)

ROUNDS:
	for round := 0; round < 99; round++ {
		fmt.Printf(";;; Check Round %d.\n", round)

		// Are we done?
		done := true
	CHECK:
		for s := range LibNeeded {
			if _, emitted := ProviderEmitted[s]; !emitted {
				done = false
				break CHECK
			}
		}
		if done {
			fmt.Printf(";;; Finished Rounds %d.\n", round)
			break ROUNDS
		}

		fmt.Printf(";;; Begin Round %d.\n", round)
	NEEDS:
		for s := range LibNeeded {
			fmt.Printf(";;; Needed %q\n", s)
			provider, ok := LibProvidedBy[s]
			if !ok {
				log.Panicf("Lib: No provider found for %q", s)
			}
			fmt.Printf(";;; Provider %q\n", provider)
			if _, emitted := ProviderEmitted[s]; emitted {
				continue NEEDS
			}
			ProviderEmitted[s] = true

			contents, ok := LibContents[provider]
			if !ok {
				log.Panicf("Lib: No contents found for %q", s)
			}

			for i, t := range contents {
				fmt.Printf("%s ;%d;\n", t, i)
				// bilbo
				if m := ArithLib.FindStringSubmatch(t); m != nil {
					DeclareLibNeeded(m[0])
				}
			}
		}
		fmt.Printf(";;; End Round %d.\n", round)
	}

	EmitPostlude()
}

func Changem(aa []*AsmLine, rules map[string]*Rule) {
	for _, a := range aa {
		if _, ok := Omit[a.Opcode]; ok {
			fmt.Printf("*** OMIT *** %q\n", a.Line)
			continue
		}

		if a.Kind == "=" {
			fmt.Printf("%s  %s  %s ; %s <longer>\n", a.Label, a.Opcode, a.Args, a.Remark)
			continue
		}

		key := Format("%s %s", a.Opcode, a.GenericArgs)
		formal := "Data"
		if strings.Contains(key, "Text") {
			formal = "Text"
		}

		if replacments, ok := rules[key]; ok {
			param := a.Parameter
			fmt.Printf("*{{{{{ %s === %q ===\n", key, a.Line)
			if a.Label != "" {
				fmt.Printf("%s:\n", key, a.Label)
			}
			for _, r := range replacments.Lines {
				fmt.Printf("  %s\n", strings.ReplaceAll(r, formal, param))
			}
			fmt.Printf("*}}}}}\n")
		} else if a.Parameter != "" {
			log.Panicf("Missing rule for this line (%s:%d) :::: %q :::: %#v", a.Filename, a.LineNum, a.Line, a)
		} else {
			fmt.Printf("%s\n", a.Line)
		}
	}
}

func Examinem(asms []*AsmLine) {
	for _, a := range asms {
		Examine1(a)
	}
}

func Examine1(a *AsmLine) {
	kind := "?"
	param := ""

	// log.Printf("@@@@@@   %v   @@@@@@", a)

	{
		if strings.HasPrefix(a.Opcode, ".") {
			kind = "."
			goto end
		}
		if _, blessed := ShortBlessed[a.Opcode]; blessed {
			kind = "="
			a.Opcode = "l" + a.Opcode
			goto end
		}
		if _, blessed := Blessed[a.Opcode]; blessed {
			kind = "!"
			goto end
		}

		m := SymbolInArgsPattern.FindStringSubmatch(a.Args)
		if m == nil {
			kind = "-"
			goto end
		}
		param = m[0]

		if _, ok := Arithmetic[param]; ok {
			Symbols[param] = &Symbol{
				Label: param,
				Area:  ".lib",
			}
		}

		symbol, ok := Symbols[param]
		if !ok {
			kind = "unknown??"
			goto end
		}

		kind = symbol.Area
		switch symbol.Area {
		case ".text", ".text.startup", ".lib", ".text.var":
			kind = "T"
		case ".data":
			kind = "D"
		case ".bss":
			kind = "D"
		default:
			kind = "?" + symbol.Area + "?"
		}
	}

end:
	arg := a.Args
	i := -1
	if param != "" {
		i = strings.Index(a.Args, param)
		AssertGE(i, 0, param, a.Args)
		addend := FindAddend(a.Args[i+len(param):])
		param += addend
		arg = Format("%s{%s:%s}%s", a.Args[0:i], kind, param, a.Args[i+len(param):])
		switch kind {
		case "T":
			a.GenericArgs = Format("%sText%s", a.Args[0:i], a.Args[i+len(param):])
		case "D":
			a.GenericArgs = Format("%sData%s", a.Args[0:i], a.Args[i+len(param):])
		default:
			// TODO // panic(kind)
		}
	}
	log.Printf("| %s | %-20s | %-10s | %-20s | %s <%s>\n", kind, a.Label, a.Opcode, arg, a.Remark, a.How)
	a.Parameter = param
	a.Kind = kind
}

var FindAddend = regexp.MustCompile(`[-+]+[0-9]+`).FindString

var Omit = map[string]bool{
	".area":  true,
	".globl": true,
}

var ShortBlessed = map[string]bool{
	"bsr": true,
	"bra": true,
	"brn": true,
	"beq": true,
	"bne": true,
	"blt": true,
	"bgt": true,
	"ble": true,
	"bge": true,
	"bhi": true,
	"blo": true,
	"bhs": true,
	"bls": true,
	"bpl": true,
	"bmi": true,
	"bcs": true,
	"bcc": true,
}

var Blessed = map[string]bool{
	"os9":  true,
	"puls": true,
	"pulu": true,
	"pshs": true,
	"pshu": true,

	"lbsr": true,

	"lbra": true,
	"lbrn": true,
	"lbeq": true,
	"lbne": true,
	"lblt": true,
	"lbgt": true,
	"lble": true,
	"lbge": true,
	"lbhi": true,
	"lblo": true,
	"lbhs": true,
	"lbls": true,
	"lbpl": true,
	"lbmi": true,
	"lbcs": true,
	"lbcc": true,
}

var Arithmetic = map[string]bool{
	"_udivhi3": true,
	"_umodhi3": true,
	"_mulhi3":  true,
	"_divhi3":  true,
	"_modhi3":  true,
	"_ashrhi3": true,
	"_lshrhi3": true,
	"_ashlhi3": true,
	"_euclid":  true,
	"_seuclid": true,
}

var ArithLib = regexp.MustCompile(`\b_((u?(div|mod|mul)|[al]sh[rl])hi3|euclid|seuclid)\b`)

func Slurp(filename string) (z []*AsmLine) {
	fd := Value(os.Open(filename))
	scanner := bufio.NewScanner(fd)
	i := 0
	for scanner.Scan() {
		i++
		line := scanner.Text()
		line = strings.TrimRight(line, " \t\r\n")

		line = strings.ReplaceAll(line, "\t", "    ")

		if m := ArithLib.FindStringSubmatch(line); m != nil {
			DeclareLibNeeded(m[0])
		}

		if m := RemarkLinePattern.FindStringSubmatch(line); m != nil {
			z = append(z, &AsmLine{
				How:      "a",
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
				How:      "b",
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
				How:      "c",
				Filename: filename,
				LineNum:  i,
				Line:     line,
				Label:    label,
			})
			continue
		}
		if label, opcode, args, remark, ok := SplitFullLine(line); ok {
			log.Printf(";nando; %q %q %q ====== %q\n", label, opcode, args, line)
			if opcode == ".area" {
				CurrentArea = FirstWord(args)
				Areas[CurrentArea] = Unit
			}
			if label != "" {
				Symbols[label] = &Symbol{
					Label: label,
					Area:  CurrentArea,
				}
			}
			z = append(z, &AsmLine{
				How:      "d",
				Filename: filename,
				LineNum:  i,
				Line:     line,
				Label:    label,
				Opcode:   strings.ToLower(opcode),
				Args:     args,
				Remark:   remark,
			})
			continue
		}
		log.Panicf("FATAL (%q:%d) gcc6809pic cannot parse %q", filename, i, line)
	}
	Check(scanner.Err(), filename)

	for i, a := range z {
		log.Printf("ASM [%d] %v", i, a)
	}

	return
}

// var EmbeddedSymbolPattern = regexp.MustCompile(`^(.*)\b([_L][A-Za-z0-9_.]*)\b(.*)$`)
// var ExtendedSymbolArgPattern = regexp.MustCompile(`^([_L][A-Za-z0-9_.]*)(\s.*)?$`)
// var ImmediateSymbolArgPattern = regexp.MustCompile(`^([#][A-Za-z_][A-Za-z0-9_.]*)(\s.*)?$`)

var EmbeddedSymbolPattern = regexp.MustCompile(`^(.*)\b([_L][A-Za-z0-9_.]*)([-+]+[0-9]+)?(.*)$`)

var FindLabel = regexp.MustCompile(`^([_L][A-Za-z0-9_.]*)([-+]+[0-9]+)?(\s.*)?$`).FindStringSubmatch
var FindOctathorpe = regexp.MustCompile(`^[#]([A-Za-z_][A-Za-z0-9_.]*)([-+]+[0-9]+)?(\s.*)?$`).FindStringSubmatch

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
		How:    "e",
		Remark: Format("*XXX* %s", a.Line),
	}
}

func EmitPrelude() {
	fmt.Println(";nando;")

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
       use os9.d
    endc

tylg                set       Prgrm+Objct
atrv                set       ReEnt+rev
rev                 set       $01
edition             set       1


size                equ       <-mem->

                    mod       eom,name,tylg,atrv,start,size

name                fcs       /<-name->/
                    fcb       edition

* Compile with gcc6809 with these:
*   -fwhole-program -fomit-frame-pointer -ffixed-y

* On entry:
*   Y =      End Params
*   D =      Size of Params
*   X = SP = Begin Params
*   U = DP = Process Memory
start:              
    pshs D,U    ; 2nd: D=params_len, 3rd: U=memory
    tfr U,Y     ; Y always points to Process Memory
    lbsr _main  ; first arg to main() in X=params
    tfr X,D     ; X has return status, but we want it in B.
    os9 F$Exit  ; should never return
zombie:         ;   but if it does:
    bra zombie  ;     F$Exit should never return.
`

const POSTLUDE = `
*** BEGIN POSTLUDE ***
    emod
eom equ *
    end
*** Generated by gcc6809pic *** fnord
`
