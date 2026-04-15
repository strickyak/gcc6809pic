package main

// Slurp a config file with [] headers

import (
	"log"
	"regexp"
	"strings"
)

var FindBlankLine = regexp.MustCompile(`^\s*$`).FindStringSubmatch
var FindCommentLine = regexp.MustCompile(`^\s*[#]`).FindStringSubmatch
var FindHeaderLine = regexp.MustCompile(`^\s*[[](.*)[]]\s*$`).FindStringSubmatch

type Rule struct {
	Key     string
	LineNum int
	Lines   []string
}

func ReadConfigFile() map[string]*Rule {
	rules := make(map[string]*Rule)
	var cd *Rule

	s := RULES_TXT
	s = strings.ReplaceAll(s, "\r\n", "\n")
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
			cd = &Rule{
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

// Each rule starts with a [...] header naming the
// generic form of the instruction to match.
// There is always one symbolic label in the header,
// which must be "Text" or "Data" depending on whether
// the label is in the read-only text portion of the module,
// or in the read-write memory of the OS9 process.
//
// The body of the rule is the text to substitute.
// Basically, any rule with "Text" needs to interpret
// that label in ",pcr" (program counter relative) mode.
// Any rule with "Data" needs to add the contents of the
// Y register to the symbol value. (The preamble of the
// program must set up the Y register, and the Y register
// must be reserved, as with the command line flag
// "-ffixed-y".   Also to free up the U register,
// use the command line flag "-fomit-frame-pointer".
const RULES_TXT = `
[ jmp Text ]
  lbra Text

[ jsr Text ]
  lbsr Text

[ ldd #Text ]
  tfr d,x
  leax Text,pcr
  tfr d,x
[ ldx #Text ]
  leax Text,pcr
[ ldu #Text ]
  leau Text,pcr

[ ldx #Text ]
  leax Text,pcr
[ ldx Text ]
  leax Text,pcr
[ ldy Text ]
  leay Text,pcr
[ ldu Text ]
  leau Text,pcr

[ ldy Text ]
  leay Text,pcr
  ldy ,y
[ ldu Text ]
  leau Text,pcr
  ldu ,u

[ leax Text ]
  leax Text,pcr
[ leay Text ]
  leay Text,pcr
[ leau Text ]
  leau Text,pcr

[ lda Text,x ]
  pshs b,x
  tfr x,d
  leax Text,pcr
  leax d,x
  lda ,x
  puls b,x
[ ldb Text,x ]
  pshs a,x
  tfr x,d
  leax Text,pcr
  leax d,x
  ldb ,x
  puls a,x
[ ldd Text,x ]
  pshs x
  tfr x,d
  leax Text,pcr
  leax d,x
  ldd ,x
  puls x

[ lda Text,u ]
  pshs b,u
  tfr u,d
  leau Text,pcr
  leau d,u
  lda ,u
  puls b,u
[ ldb Text,u ]
  pshs a,u
  tfr u,d
  leau Text,pcr
  leau d,u
  ldb ,u
  puls a,u
[ ldd Text,u ]
  pshs u
  tfr u,d
  leau Text,pcr
  leau d,u
  ldd ,u
  puls u

[ jmp [Text,x] ]
  pshs u,x   ; u is a place holder
  leax Text,pcr
  ldx ,x
  stx 2,s    ; where U was put
  puls x,pc  ; restores X and branches to 2,s

[ lda [Data] ]
  pshs x
  leax Data,y
  lda ,x
  puls x
[ ldd [Data] ]
  pshs x
  leax Data,y
  ldd ,x
  puls x
[ ldb [Data] ]
  pshs x
  leax Data,y
  ldb ,x
  puls x
[ ldx [Data] ]
  leax Data,y
  ldx ,x

[ leax Data,x ]
  exg d,y
  leax Data,x
  leax d,x
  exg d,y
[ leau Data,u ]
  exg d,y
  leau Data,u
  leau d,u
  exg d,y

[ addd Data ]
  pshs x
  ldx Data,y
  leax d,x
  puls x

[ addd #Data ]
  pshs x
  leax Data,y
  leax d,x
  tfr x,d
  puls x

[ ldd #Data ]
  exg d,x
  leax Data,y
  exg d,x
[ ldx #Data ]
  leax Data,y
[ ldu #Data ]
  leau Data,y

[ lda Data ]
  lda Data,y
[ ldb Data ]
  ldb Data,y
[ ldd Data ]
  ldd Data,y
[ ldx Data ]
  ldx Data,y
[ ldu Data ]
  ldu Data,y

[ sta Data ]
  sta Data,y
[ stb Data ]
  stb Data,y
[ std Data ]
  std Data,y
[ stx Data ]
  stx Data,y
[ stu Data ]
  stu Data,y

[ std Data,x ]
  pshs x,d
  tfr y,d
  leax d,x
  puls d
  std Data,x
  puls x
[ ldd Data,x ]
  pshs x
  tfr y,d
  leax d,x
  ldd Data,x
  puls x

[ lda Data,x ]
  pshs b,x
  tfr y,d
  leax d,x
  lda Data,x
  puls b,x
[ lda Data,u ]
  pshs b,u
  tfr y,d
  leau d,u
  lda Data,u
  puls b,u

[ ldb Data,x ]
  pshs a,x
  tfr y,d
  leax d,x
  ldb Data,x
  puls a,x
[ ldb Data,u ]
  pshs a,u
  tfr y,d
  leau d,u
  ldb Data,u
  puls a,u

[ anda Data,x ]
  pshs d,x
  tfr y,d
  leax d,x
  puls d
  anda Data,x
  puls x
[ andb Data,x ]
  pshs d,x
  tfr y,d
  leax d,x
  puls d
  andb Data,x
  puls x

[ cmpa Data ]
  cmpa Data,y
[ cmpb Data ]
  cmpb Data,y
[ cmpd Data ]
  cmpd Data,y
[ cmpx Data ]
  cmpx Data,y
[ cmpu Data ]
  cmpu Data,y

[ cmpu #Text ]
  pshs x
  leax Text,pcr
  pshs x
  cmpu ,s
  puls x
  puls x

[ cmpd #Text ]
  pshs x
  leax Text,pcr
  pshs x
  cmpd ,s
  puls x
  puls x


[ clr Data ]
  clr Data,y
[ clr Data,x ]
  pshs u
  exg x,d
  leau Data,y
  clr d,u
  exg x,d
  puls u
[ clr Data,u ]
  pshs x
  exg u,d
  leax Data,y
  clr d,x
  exg u,d
  puls x

[ stb Data,x ]
  pshs a,x
  pshs b
  tfr y,d
  leax d,x
  puls b
  stb Data,x
  puls a,x
[ stb Data,u ]
  pshs a,u
  pshs b
  tfr y,d
  leau d,u
  puls b
  stb Data,u
  puls a,u

[ leau Data,x ]
  exg x,d
  leau Data,y
  leau d,u
  exg x,d
`
