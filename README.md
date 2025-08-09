# gcc6809pic

Fix gcc6809(4.6.4) Position-Independent Code with ASM preprocessor

We have a version of gcc4.6.4 with a Motorola 6809 back end.

It has bugs and limitations, but when it works, it can produce small
fast code much better than "naive" compilers that have local and peephole
optimiztions but no global optimizations.

Unfortunately it cannot be used for OS9/NitrOS9 modules, which require
`-fpic`, which is severely broken on this version of gcc.

So this project attempts to fix it, by compiling without `-fpic` and then
making textual edits to the ASM output of gcc (sort of like a peephole
optimizer does) before compiling that with LWASM.

There will probably be some C constructs that we just can't or don't
handle.  But with some discipline in how you use C, it can work.

## Work In Progress

This only works for one program `a.c`.

Global/Static variables definitely do not work at all (yet).

Many addressing modes (like `  ldb _buf+1`) do not work (yet).

I'm using `Ubuntu 24.04.2 LTS` on `x86_64`.

## Where is the gcc?

https://github.com/strickyak/coco-shelf can get you started.

## Roadmap

* Finish `.text` patterns

* Do `.data` & `.bss` patterns (for NitrOS9, Level I or II)

* Use `-fwhole-program` for global optimiztions,
  but that requires an "#include"-only approach, rather than multiple
  program modules.

* Smartly include needed libraries, so that `lwlink`
  is never needed, and the whole program appears to `lwasm --format=os9`
  as if it is all in just one .asm file.
