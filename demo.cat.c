/*

hd _run

00000300  7e 7e 0d 0a 4e 69 74 72  4f 53 2d 39 2f 36 38 30  |~~..NitrOS-9/680|
00000310  39 20 4c 65 76 65 6c 20  31 20 56 49 2e 37 2e 3f  |9 Level 1 VI.7.?|
00000320  0d 0a 52 61 64 69 6f 20  53 68 61 63 6b 20 43 6f  |..Radio Shack Co|
00000330  6c 6f 72 20 43 6f 6d 70  75 74 65 72 0d 0a 28 43  |lor Computer..(C|
00000340  29 20 32 30 31 34 20 54  68 65 20 4e 69 74 72 4f  |) 2014 The NitrO|
00000350  53 2d 39 20 50 72 6f 6a  65 63 74 0d 0a 54 75 65  |S-9 Project..Tue|
00000360  20 4a 75 6c 20 31 35 20  32 31 3a 33 38 3a 32 35  | Jul 15 21:38:25|
00000370  20 32 30 32 35 0d 0a 68  74 74 70 3a 2f 2f 77 77  | 2025..http://ww|
00000380  77 2e 6e 69 74 72 6f 73  39 2e 6f 72 67 0d 0a 0a  |w.nitros9.org...|
00000390  66 64 3d 33 0d 0a 61 63  74 75 61 6c 5f 6e 75 6d  |fd=3..actual_num|
000003a0  5f 62 79 74 65 73 5f 6f  75 74 3d 39 0d 0a 64 65  |_bytes_out=9..de|
000003b0  6d 6f 2e 63 61 74 0d 77  72 69 74 74 65 6e 5f 61  |mo.cat.written_a|
000003c0  63 74 75 61 6c 5f 6e 75  6d 5f 62 79 74 65 73 5f  |ctual_num_bytes_|
000003d0  6f 75 74 3d 39 0d 0a 0d  0a 53 68 65 6c 6c 0d 0a  |out=9....Shell..|
000003e0  0d 0a 4f 53 39 3a                                 |..OS9:|
000003e6


OS9:demo.cat ! dump

Address   0 1  2 3  4 5  6 7  8 9  A B  C D  E F  0 2 4 6 8 A C E
-------- ---- ---- ---- ---- ---- ---- ---- ----  ----------------
00000000 6664 3D33 0A61 6374 7561 6C5F 6E75 6D5F  fd=3.actual_num_
00000010 6279 7465 735F 6F75 743D 390A 6465 6D6F  bytes_out=9.demo
00000020 2E63 6174 0D77 7269 7474 656E 5F61 6374  .cat.written_act
00000030 7561 6C5F 6E75 6D5F 6279 7465 735F 6F75  ual_num_bytes_ou
00000040 743D 390A                                t=9.

*/

#include "demo.types.h"
#include "demo.printf.h"
#include "os9/_generated_os9api_for_gcc.h"
#include "os9/_generated_os9api_for_gcc.c"

char buf[256];

int main(void) {
    errnum e;
    byte fd;

    {
        word junk;
        e = GccOs9Open(/*access_mode*/1, (word)"startup", &fd, &junk);
        if (e) {
            printf("cannot open startup ($%x=%d.)\n", e, e);
            goto Exit;
        }
        printf("fd=%d\n", fd);
    }
    word actual_num_bytes_out;
    {
        e = GccOs9Read(fd, (word)buf, sizeof buf, &actual_num_bytes_out);
        if (e) {
            printf("cannot read startup ($%x=%d.)\n", e, e);
            goto Exit;
        }
        printf("actual_num_bytes_out=%d\n", actual_num_bytes_out);
    }

    word written_actual_num_bytes_out;
    {
        e = GccOs9Write(1, (word)buf, actual_num_bytes_out, &written_actual_num_bytes_out);
        if (e) {
            printf("cannot write ($%x=%d.)\n", e, e);
            goto Exit;
        }
        printf("written_actual_num_bytes_out=%d\n", written_actual_num_bytes_out);
    }

Exit:
    e = GccOs9Exit(e);
    printf("why did Exit fail ($%x=%d.)", e, e);
}
