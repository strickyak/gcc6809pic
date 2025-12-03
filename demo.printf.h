#ifndef _DEMO_PRINTF_H_
#define _DEMO_PRINTF_H_

#include "demo.types.h"

typedef void (*emitter)(byte);

static void PutChar(byte c) {
    volatile byte c2 = c;
    volatile byte* pointer = &c2;

    asm volatile (
        "*** PutChar *** \n"
        " ldx %[pointer] ; points to the byte \n"
        " pshs y \n"
        " lda #1 ; path 1 is stdout \n"
        " ldy #1 ; just 1 byte \n"
        " os9 I$WritLn \n"
        " puls y \n"
        : // no outputs
        : [pointer] "m" (pointer) // inputs
        : "d", "x" // clobbers
   );
}

// Avoid `%` and `/` with denominator 10.
static byte PDivMod10(word x, word* out_quotient) {  // returns residue
  word quotient = 0;
  while (x >= 10000) x -= 10000, quotient += 1000;
  while (x >= 1000) x -= 1000, quotient += 100;
  while (x >= 100) x -= 100, quotient += 10;
  while (x >= 10) x -= 10, quotient++;
  *out_quotient = quotient;
  return (byte)x;
}

static void FormatUint(emitter fn, word x) {
    if (x > 9) {
        word quotient = 0;
        x = PDivMod10(x, &quotient);
        FormatUint(fn, quotient);
    }
    fn('0' + x);
}

static void FormatHex(emitter fn, word x) {
    if (x > 15) {
        word quotient = 0;
        FormatHex(fn, x>>4);
        x = x & 15;
    }
    fn( (x>0) ? 'a' + x - 10 : '0' + x);
}

static void FormatInt(emitter fn, int x) {
    if (x < 0) {
        fn('-');
        x = (-x);
    }
    FormatUint(fn, x);
}

// the heart of printf, like vprintf with a fn parameter.
static void Format(emitter fn, const char* fmt, va_list ap) {
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

static int printf(const char* fmt, ...) {
    va_list ap;
    va_start(ap, fmt);
    Format(PutChar, fmt, ap);
    va_end(ap);
    return 1;  // bogus return value
}

#endif // _DEMO_PRINTF_H_
