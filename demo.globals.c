register int _data_ asm ("y");

#include "demo.types.h"
#include "demo.printf.h"

int aaa;
word bbb;
byte ccc;

void run(void) {
    ccc = aaa + bbb;
}

volatile int foo;
volatile byte bar;

const char TWELVE[] = "Twelve";

int main(void) {
    foo = 12345;
    aaa = 5;
    bbb = 7;
    if (foo == 12345) run();
    printf("%s = %d\n", TWELVE, ccc);
    ++foo;
}
