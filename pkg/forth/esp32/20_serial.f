\ Copyright 2024-2025 Blake Felt blake.w.felt@gmail.com
\ This Source Code Form is subject to the terms of the Mozilla Public
\ License, v. 2.0. If a copy of the MPL was not distributed with this
\ file, You can obtain one at https://mozilla.org/MPL/2.0/.

\ Create the assembly to bitbang serial writes.
\ Takes in a pin number and wait time, then parses
\ the next word. Creates a definition for that word
\ that can be used to write one character over serial.
\
\ This is in assembly because the forth timings
\ may change later with optimization passes, assembly
\ is consistent. It's also difficult to tune high baud
\ rates in forth without taking a lot of space.
: SERIAL.WRITE_CREATE.BUILDER ( pin wait-time -- )
    2>R \ put arguments on return stack while building
    C" ld r0, r3, 0\n" \ load the character
    C" add r3, r3, 1\n" \ decrement the return stack
    C" or r0, r0, 0x100\n" \ set the end bit
    C" lsh r0, r0, 1\n" \ set the start bit
    C" stage_rst\n" \ reset the stage counter
    C" __serial_write_0_" 1 RPICK C" _" 0 RPICK C" :\n"
        C" and r1, r0, 1\n" \ get the lowest bit, 4 cycles
        C" jump __serial_write_1_" 1 RPICK C" _" 0 RPICK C" , eq\n" \ 4 cycles
            \ set high, 12 cycles
            RTCIO_RTC_GPIO_OUT_W1TS_REG
            RTCIO_RTC_GPIO_OUT_DATA_W1TS_S 1 RPICK +
            1 1
            WRITE_RTC_REG.BUILDER >C
            C" jump __serial_write_2_" 1 RPICK C" _" 0 RPICK C" \n" \ 4 cycles
        C" __serial_write_1_" 1 RPICK C" _" 0 RPICK C" :\n"
            \ set low, 12 cycles
            RTCIO_RTC_GPIO_OUT_W1TC_REG
            RTCIO_RTC_GPIO_OUT_DATA_W1TC_S 1 RPICK +
            1 1
            WRITE_RTC_REG.BUILDER >C
            \ jump so it takes the same time, 4 cycles
            C" jump __serial_write_2_" 1 RPICK C" _" 0 RPICK C" \n"
        C" __serial_write_2_" 1 RPICK C" _" 0 RPICK C" :\n"
        C" wait " R@ C" \n" \ wait for wait-time ticks, 6+n cycles
        C" rsh r0, r0, 1\n" \ go to next bit, 4 cycles
        \ loop 8 times
        C" stage_inc 1\n" \ 4 cycles
        C" jumps __serial_write_0_" 1 RPICK C" _" 0 RPICK C" , 10, lt\n" \ 4 cycles
    \ time from reg_wr to reg_wr (only 1 reg_wr):
    \ 12+4 + 6+n+4+4+4 + 4+4
    \ = 42+n cycles
    46 C> C> + + \ get the total number of inputs
    2R> 2DROP \ clean up the return stack
;

: SERIAL.WRITE_CREATE ( pin wait-time "<spaces>name" -- )
    2DUP >R >R \ dup the input
    \ token threaded
    SERIAL.WRITE_CREATE.BUILDER
    \ subroutine threaded
    R> R>
    SERIAL.WRITE_CREATE.BUILDER
    ASSEMBLY-BOTH
    TOKEN_NEXT_SKIP_R2 LAST SET-ULP-ASM-NEXT
;

\ these were found at 21 C with a logic analyzer
\ 0 CONSTANT SERIAL.WRITE_4800_BAUD
874 CONSTANT SERIAL.WRITE_9600_BAUD
\ 0 CONSTANT SERIAL.WRITE_19200_BAUD
\ 0 CONSTANT SERIAL.WRITE_38400_BAUD
\ 0 CONSTANT SERIAL.WRITE_57600_BAUD
34  CONSTANT SERIAL.WRITE_115200_BAUD \ fastest we can go with this algorithm
