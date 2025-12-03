#include "demo.types.h"
#include "demo.printf.h"

int main(void) {
    for (char* s = "Hello World!\n"; *s; s++) PutChar(*s);
}
