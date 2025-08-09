V3=/home/strick/modoc/coco-shelf/tfr9/v3
N9=/home/strick/modoc/coco-shelf/nitros9
BORGES=/home/strick/borges
#PRAGMA=--pragma=pcaspcr,nosymbolcase,condundefzero,undefextern,dollarnotlocal,noforwardrefmax,cescapes
PRAGMA=--pragma=condundefzero,undefextern,dollarnotlocal,noforwardrefmax,cescapes

all: noname.mod

gcc6809pic: gcc6809pic.go util.go
	go build gcc6809pic.go util.go 

a.s: a.c
	gcc6809 -std=gnu99 -Os -fomit-frame-pointer -fwhole-program -S $<

noname.asm: a.s gcc6809pic
	./gcc6809pic $< > $@

noname.mod: noname.asm
	lwasm $(PRAGMA) --format=os9 $< -o'$@' -I$(N9)/level1/coco1 -I$(N9)/level1/coco1/defs
	go run $(V3)/borges-saver/borges-saver.go --outdir=$(BORGES) .
	os9 copy -r noname.mod $(V3)/generated/level1.dsk,cmds/noname
	os9 attr -e -pe  $(V3)/generated/level1.dsk,cmds/noname
	os9 ident $(V3)/generated/level1.dsk,cmds/noname

ci:
	ci-l *.c *.go Makefile

clean:
	cp -f *.s *.asm *.mod *.mod.list *.mod.map gcc6809pic /tmp/
	rm -f *.s *.asm *.mod *.mod.list *.mod.map gcc6809pic
