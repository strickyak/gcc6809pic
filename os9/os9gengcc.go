package main

import (
	"bytes"
    "flag"
	. "fmt"
	"io"
	"io/ioutil"
	"log"
)

var OmitFramePointer = flag.Bool("fomit-frame-pointer", false, "If you set -fomit-frame-pointer on GCC, you must also set it here.")

// assumes all args are on stack.
func (c *Os9ApiCall) FormatArgsForGcc() string {
	var bb bytes.Buffer
	if c.A != "" {
		Fprintf(&bb, "    /* A */ volatile byte %s,\n", c.A)
	}
	if c.B != "" {
		Fprintf(&bb, "    /* B */ volatile byte %s,\n", c.B)
	}
	if c.D != "" {
		Fprintf(&bb, "    /* D */ volatile word %s,\n", c.D)
	}
	if c.X != "" {
		Fprintf(&bb, "    /* X */ volatile word %s,\n", c.X)
	}
	if c.Y != "" {
		Fprintf(&bb, "    /* Y */ volatile word %s,\n", c.Y)
	}
	if c.U != "" {
		Fprintf(&bb, "    /* U */ volatile word %s,\n", c.U)
	}
	if c.RA != "" {
		Fprintf(&bb, "    /* RA */ volatile byte* %s_out,\n", c.RA)
	}
	if c.RB != "" {
		Fprintf(&bb, "    /* RB */ volatile byte* %s_out,\n", c.RB)
	}
	if c.RD != "" {
		Fprintf(&bb, "    /* RD */ volatile word* %s_out,\n", c.RD)
	}
	if c.RX != "" {
		Fprintf(&bb, "    /* RX */ volatile word* %s_out,\n", c.RX)
	}
	if c.RY != "" {
		Fprintf(&bb, "    /* RY */ volatile word* %s_out,\n", c.RY)
	}
	if c.RU != "" {
		Fprintf(&bb, "    /* RU */ volatile word* %s_out,\n", c.RU)
	}
	s := bb.String()
	n := len(s)
	if n > 0 {
		s = s[:n-2] + "\n" // remove final comma
	}
	return s
}

func PrintCForGcc(c *Os9ApiCall, w io.Writer) {
	P := func(format string, args ...any) {
		Fprintf(w, format+"\n", args...)
	}

    P("")
    P("/////////// %s // %s", c.Name, c.Desc)
    // __attribute__((externally_visible))
    // __attribute__((noinline))
    // Problem: when inlining, it did not issue PSHS Y PULS Y,
    // and Y got clobbered.  So let's try attribute `noinline`.
	P("__attribute__((noinline))")
	P("errnum GccOs9%s  (\n", c.Name[2:])
	P("%s) {\n", c.FormatArgsForGcc())
    P("  errnum _err_ = 0;\n");

	if c.A != "" {
		P("    volatile byte vol_a = (byte)%s;", c.A)
	}
	if c.B != "" {
		P("    volatile byte vol_b = (byte)%s;", c.B)
	}
	if c.D != "" {
		P("    volatile word vol_d = (word)%s;", c.D)
	}
	if c.X != "" {
		P("    volatile word vol_x = (word)%s;", c.X)
	}
	if c.Y != "" {
		P("    volatile word vol_y = (word)%s;", c.Y)
	}
	if c.U != "" {
		P("    volatile word vol_u = (word)%s;", c.U)
	}

	if c.RA != "" {
		P("    volatile word vol_ra = (word)%s_out;", c.RA)
	}
	if c.RB != "" {
		P("    volatile word vol_rb = (word)%s_out;", c.RB)
	}
	if c.RD != "" {
		P("    volatile word vol_rd = (word)%s_out;", c.RD)
	}
	if c.RX != "" {
		P("    volatile word vol_rx = (word)%s_out;", c.RX)
	}
	if c.RY != "" {
		P("    volatile word vol_ry = (word)%s_out;", c.RY)
	}
	if c.RU != "" {
		P("    volatile word vol_ru = (word)%s_out;", c.RU)
	}
	P("    volatile word savedY;")
    if *OmitFramePointer {
	    P("    volatile word savedU;")
    }

    P(`    asm volatile ("\n"`);
    P(`    "  sty %%[savedY]\n"`)
    if *OmitFramePointer {
        P(`    "  stu %%[savedU]\n"`)
    }
	if c.A != "" {
		P(`    "  lda %%[vol_a]\n"`)
	}
	if c.B != "" {
		P(`    "  ldb %%[vol_b]\n"`)
	}
	if c.D != "" {
		P(`    "  ldd %%[vol_d]\n"`)
	}
	if c.X != "" {
		P(`    "  ldx %%[vol_x]\n"`)
	}
	if c.Y != "" {
		P(`    "  ldy %%[vol_y]\n"`)
	}
	if c.U != "" {
		P(`    "  ldu %%[vol_u]\n"`)
	}

	P(`    "  os9 $%02x ; %s ; %s\n"`, c.Number, c.Name, c.Desc)
    P(`    "  bcc @OK \n"`)
    P(`    "  stb %%[_err_]\n"`)
	P(`    "  bra @END \n"`)
	P(`    "@OK:\n"`)

	if c.RA != "" {
		P(`    "  sta [%%[vol_ra]]\n"`)
	}
	if c.RB != "" {
		P(`    "  stb [%%[vol_rb]]\n"`)
	}
	if c.RD != "" {
		P(`    "  std [%%[vol_rd]]\n"`)
	}
	if c.RX != "" {
		P(`    "  stx [%%[vol_rx]]\n"`)
	}
	if c.RY != "" {
		P(`    "  sty [%%[vol_ry]]\n"`)
	}
	if c.RU != "" {
		P(`    "  stu [%%[vol_ru]]\n"`)
	}

	P(`    "@END:\n"`)
    if *OmitFramePointer {
        P(`    "  ldu %%[savedU]\n"`)
    }
    P(`    "  ldy %%[savedY]\n"`)

    P(`  : // outputs`)

	if c.RA != "" {
		P(`    [vol_ra] "=m" (vol_ra),`)
	}
	if c.RB != "" {
		P(`    [vol_rb] "=m" (vol_rb),`)
	}
	if c.RD != "" {
		P(`    [vol_rd] "=m" (vol_rd),`)
	}
	if c.RX != "" {
		P(`    [vol_rx] "=m" (vol_rx),`)
	}
	if c.RY != "" {
		P(`    [vol_ry] "=m" (vol_ry),`)
	}
	if c.RU != "" {
		P(`    [vol_ru] "=m" (vol_ru),`)
	}
    if *OmitFramePointer {
		P(`    [savedY] "=m" (savedY),`)
    }
	P(`    [savedU] "=m" (savedU),`)
	P(`    [_err_] "=m" (_err_)`)


    P("  : // inputs")

    comma := " "
	if c.A != "" {
		P(`    %s [vol_a] "m" (vol_a)`, comma)
        comma = ","
	}
	if c.B != "" {
		P(`    %s [vol_b] "m" (vol_b)`, comma)
        comma = ","
	}
	if c.D != "" {
		P(`    %s [vol_d] "m" (vol_d)`, comma)
        comma = ","
	}
	if c.X != "" {
		P(`    %s [vol_x] "m" (vol_x)`, comma)
        comma = ","
	}
	if c.Y != "" {
		P(`    %s [vol_y] "m" (vol_y)`, comma)
        comma = ","
	}
	if c.U != "" {
		P(`    %s [vol_u] "m" (vol_u)`, comma)
        comma = ","
	}



    P(`  : "d", "x", "y", "u" // clobbers`)
    P(`  );`)
    P(`   return _err_;`)
    P(`};`)
}

func PrintCallsForGcc() {
	var gen_hdr bytes.Buffer
	Fprintf(&gen_hdr, "#ifndef _GEN_HDR_FOR_GCC_\n")
	Fprintf(&gen_hdr, "#define _GEN_HDR_FOR_GCC_\n")
	Fprintf(&gen_hdr, "#include \"types_gcc6809.h\"\n")

	for _, c := range Os9ApiCalls {
		Fprintf(&gen_hdr, "\nextern errnum GccOs9%s(\n", c.Name[2:])
		Fprintf(&gen_hdr, "%s);\n", c.FormatArgsForGcc())
	}
	Fprintf(&gen_hdr, "#endif\n")

	const hdr_filename = "_generated_os9api_for_gcc.h"
	err2 := ioutil.WriteFile(hdr_filename, gen_hdr.Bytes(), 0777)
	if err2 != nil {
		log.Fatalf("Cannot write %q: %v", hdr_filename, err2)
	}

	/////////////////////////////

	const c_filename = "_generated_os9api_for_gcc.c"

	var gen_asm bytes.Buffer

    Fprintln(&gen_asm, `#include "types_gcc6809.h"`)
    Fprintln(&gen_asm, `#include "_generated_os9api_for_gcc.h"`)
    Fprintln(&gen_asm, ``)
	for _, c := range Os9ApiCalls {
		PrintCForGcc(c, &gen_asm)
	}

	err := ioutil.WriteFile(c_filename, gen_asm.Bytes(), 0777)
	if err != nil {
		log.Fatalf("Cannot write %q: %v", c_filename, err)
	}
}
