#include "demo.types.h"
#include "demo.printf.h"

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
