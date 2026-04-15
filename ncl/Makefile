V3=/home/strick/modoc/coco-shelf/tfr9/v3
N9=/home/strick/modoc/coco-shelf/nitros9
BORGES=/home/strick/borges
XXX_PRAGMA=--pragma=condundefzero,undefextern,dollarnotlocal,noforwardrefmax,cescapes
PRAGMA=--pragma=condundefzero,undefextern,noforwardrefmax,cescapes,dollarlocal

all: ncl.os9

os9-dir:
	make -C ../os9

%.os9: %.c ../gcc6809pic.go ../util.go os9-dir
	:
	gcc6809 -std=gnu99 -Os -ffixed-y -fomit-frame-pointer -fwhole-program -msoft-reg-count=4 -S -I.. -I/home/strick/modoc/coco-shelf/frobio/ -D'BASENAME'='"ncl"' $*.c
	:
	go run ../../gcc6809pic/ --name='$*' -- $*.s > $*.asm
	:
	: : : : lwasm $(PRAGMA) --format=os9 $*.asm -o'$@' -I$(N9)/level1/coco1 -I$(N9)/level1/coco1/defs --list=$*.os9.list --map=$*.os9.map
	lwasm $(PRAGMA) --format=os9 $*.asm -o'$@' -I. -DLevel=1 -DH6309=0 --list=$*.os9.list --map=$*.os9.map
	:
	go run $(V3)/borges-saver/borges-saver.go --outdir=$(BORGES) .
	os9 copy -r $*.os9 $(V3)/generated/level1.dsk,cmds/$*
	os9 attr -q -e -pe  $(V3)/generated/level1.dsk,cmds/$*
	os9 ident $(V3)/generated/level1.dsk,cmds/$*

ci:
	ci-l *.h *.c *.go Makefile

clean:
	-cp -f *.s *.asm *.os9 *.os9.list *.os9.map /tmp/
	rm -f *.s *.asm *.os9 *.os9.list *.os9.map

# HINTS
# ncl2$ lwasm -f os9  -oncl.os9  --list=ncl.os9.list  ncl.asm -I. -DLevel=1 -DH6309=0
# + exec lwasm.orig -f os9 -oncl.os9 --list=ncl.os9.list ncl.asm -I. -DLevel=1 -DH6309=0 --map=ncl.os9.map --list=ncl.os9.list
#OLD: : : : lwasm $(PRAGMA) --format=os9 $*.asm -o'$@' -I$(N9)/level1/coco1 -I$(N9)/level1/coco1/defs --list=$*.os9.list --map=$*.os9.map
