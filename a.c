// Simple first experimental gcc6809 C code,
// to turn into a position-independant OS9 module.

typedef unsigned char bool;
typedef unsigned char byte;
typedef unsigned int word;
typedef int (*binop)(int a, int b);

#include <stdarg.h>

typedef void (*emitter)(byte);

void PutChar(byte c) {
    volatile byte c2 = c;
    volatile byte* pointer = &c2;

    asm volatile (
        "*** PutChar *** \n"
        " lda #1 ; path 1 is stdout \n"
        " ldy #1 ; just 1 byte \n"
        " ldx %[pointer] ; points to the byte \n"
        " os9 I$WritLn \n"
        : // no outputs
        : [pointer] "m" (pointer) // inputs
        : "d", "y", "x" // clobbers
   );
}

// Avoid `%` and `/` with denominator 10.
byte PDivMod10(word x, word* out_quotient) {  // returns residue
  word quotient = 0;
  while (x >= 10000) x -= 10000, quotient += 1000;
  while (x >= 1000) x -= 1000, quotient += 100;
  while (x >= 100) x -= 100, quotient += 10;
  while (x >= 10) x -= 10, quotient++;
  *out_quotient = quotient;
  return (byte)x;
}

void FormatUint(emitter fn, word x) {
    if (x > 9) {
        word quotient = 0;
        x = PDivMod10(x, &quotient);
        FormatUint(fn, quotient);
    }
    fn('0' + x);
}

void FormatInt(emitter fn, int x) {
    if (x < 0) {
        fn('-');
        x = (-x);
    }
    FormatUint(fn, x);
}

// the heart of printf, like vprintf with a fn parameter.
void Format(emitter fn, const char* fmt, va_list ap) {
    for (const char* s = fmt; *s; s++) {
        if (*s != '%') {
            fn(*s);
            continue;
        }

        s++;
        switch (*s) {
            case 'd': {
                int x;
                x = va_arg(ap, int);
                FormatInt(fn, x);
            }
            break;
            case 'u': {
                word x;
                x = va_arg(ap, word);
                FormatUint(fn, x);
            }
            break;
            case 's': {
                char* x;
                x = va_arg(ap, char*);
                while (*x) {
                    fn(*x++);
                }
            }
            break;
            default: {
                fn('?');
                fn('%');
                fn(*s);
                fn('?');
            }
        }
    }
}

int printf(const char* fmt, ...) {
    va_list ap;
    va_start(ap, fmt);
    Format(PutChar, fmt, ap);
    va_end(ap);
    return 1;  // bogus return value
}

int add(int a, int b) { return a+b; }
int sub(int a, int b) { return a-b; }

int fib(int x, binop fn) {
    if (x < 2) return x;
    return fn( fib(x-1, fn) , fib(x-2, fn) );
}

volatile int foo = 1;

int main(void) {
    binop fn = (foo) ? add : sub;  // will be `add`, but gcc cannot assume that.
    for (int i=0; i<20; i++) {
        printf("fib(%d) = %d\n", i, fib(i, fn));
    }
}
