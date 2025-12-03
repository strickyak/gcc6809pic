V3=/home/strick/modoc/coco-shelf/tfr9/v3
N9=/home/strick/modoc/coco-shelf/nitros9
BORGES=/home/strick/borges
PRAGMA=--pragma=condundefzero,undefextern,dollarnotlocal,noforwardrefmax,cescapes

all: ncl.os9

os9-dir:
	make -C ../os9

%.os9: %.c ../gcc6809pic.go ../util.go os9-dir
	gcc6809 -std=gnu99 -Os -ffixed-y -fomit-frame-pointer -fwhole-program -msoft-reg-count=4 -S -I.. -I/home/strick/modoc/coco-shelf/frobio/ -D'BASENAME'='"ncl"' $*.c
	go run ../gcc6809pic.go ../util.go ../lib1.go --name='$*' -- $*.s > $*.asm
	lwasm $(PRAGMA) --format=os9 $*.asm -o'$@' -I$(N9)/level1/coco1 -I$(N9)/level1/coco1/defs --list=$*.os9.list --map=$*.os9.map
	go run $(V3)/borges-saver/borges-saver.go --outdir=$(BORGES) .
	os9 copy -r $*.os9 $(V3)/generated/level1.dsk,cmds/$*
	os9 attr -q -e -pe  $(V3)/generated/level1.dsk,cmds/$*
	os9 ident $(V3)/generated/level1.dsk,cmds/$*

ci:
	ci-l *.h *.c *.go Makefile

clean:
	-cp -f *.s *.asm *.os9 *.os9.list *.os9.map /tmp/
	rm -f *.s *.asm *.os9 *.os9.list *.os9.map
