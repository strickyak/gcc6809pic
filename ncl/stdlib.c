typedef unsigned int size_t;


void EnableIrqsCounting() {}
void DisableIrqsCounting() {}

void Puts(const char* s) {
  while (*s) {
    word n;
    GccOs9WritLn(1, s, 1, &n);
    s++;
  }
}
void StderrPuts(const char* s) {
  while (*s) {
    word n;
    GccOs9WritLn(2, s, 1, &n);
    s++;
  }
}

void Panic(const char* s) {
  StderrPuts(s);
  asm volatile("panicked: bra panicked");
}

void abort(void) {
  Panic(" *ABORT* *LOOP*\n");
  while(1){}
}

void* memcpy(void* dest, const void* src, size_t n) {
   char* d = dest;
   const char* s = src;
   for (size_t i=0; i<n; i++) {
     d[i] = s[i];
   }
   return d;
}

void* memset(void* dest, int ch, size_t n) {
   char* d = dest;
   for (size_t i=0; i<n; i++) {
     d[i] = (char)ch;
   }
   return d;
}

char* strcat(char* dest, const char* src) {
   char* d = dest;
   const char* s = src;
   while (*d) d++;
   while (*s) {
     *d++ = *s++;
   }
   *d = '\0';
   return dest;
}

char *strcpy(char *dest, const char *src) {
   char* d = dest;
   const char* s = src;
   while (*s) {
     *d++ = *s++;
   }
   *d = '\0';
   return dest;
}

char *strncpy(char *dest, const char *src, size_t n) {
   char* d = dest;
   const char* s = src;
   while (*s) {
     *d++ = *s++;
     n--;
     if (n==0) return dest;
   }
   *d = '\0';
   return dest;
}

size_t strlen(const char *str) {
    const char* s = str;
    while (*s) {
        s++;
    }
    return s - str;
}

int strcmp(const char* a, const char* b) {
  while (*a == *b) a++, b++;
  if (*a < *b) return -1;
  if (*a > *b) return +1;
  return 0;
}

int atoi(const char* s) {
  int z = 0;
  while ('0' <= *s && *s <= '9') {
    z = 10*z + (*s - '0');
  }
  return z;
}

// // void PutHex(byte c, word x) {
    // // // TODO
// // }

char* strdup(const char* s) {
    word n = strlen(s);
    char* p = malloc(n+1);
    memcpy(p, s, n+1);
    return p;
}
