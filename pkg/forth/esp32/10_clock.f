\ Copyright 2024-2025 Blake Felt blake.w.felt@gmail.com
\ This Source Code Form is subject to the terms of the Mozilla Public
\ License, v. 2.0. If a copy of the MPL was not distributed with this
\ file, You can obtain one at https://mozilla.org/MPL/2.0/.

\ Create word RTC_CLOCK to read the lower 32 bits of the rtc clock.
\ tell rtc timer to update
RTC_CNTL_TIME_UPDATE_REG RTC_CNTL_TIME_UPDATE_S 1 1
WRITE_RTC_REG.BUILDER >C
STRING" sub r3, r3, 2\n"
STRING" __RTC_CLOCK.0:\n"
    \ read to see if updated
    RTC_CNTL_TIME_UPDATE_REG RTC_CNTL_TIME_VALID_S 1
    READ_RTC_REG.BUILDER >C
    \ loop if not
    STRING" jumpr  __RTC_CLOCK.0, 1, lt\n"
\ read 0..15
RTC_CNTL_TIME0_REG 0 16
READ_RTC_REG.BUILDER >C
\ store result
STRING" st r0, r3, 1\n"
\ read 16..31
RTC_CNTL_TIME0_REG 16 16
READ_RTC_REG.BUILDER >C
\ store result and exit
STRING" st r0, r3, 0"
5 C> C> C> C> + + + + \ add up the strings and the built instructions
ASSEMBLY RTC_CLOCK \ create RTC_CLOCK
TOKEN_NEXT_SKIP_LOAD LAST SET-ULP-ASM-NEXT

\ delay for d rtc_slow ticks
: RTC_CLOCK_DELAY ( d -- )
    RTC_CLOCK ( d cycles ) \ read the current cycles
    BEGIN
        2OVER 2OVER ( d cycles d cycles )
        RTC_CLOCK ( d cycles d cycles new )
        2SWAP D- ( d cycles d diff )
        DU< ( d cycles bool )
    UNTIL
    2DROP 2DROP \ clean up stack
;

: BUSY_DELAY.BUILDER
    C" ld r0, r3, 0\n"
    C" jumpr __busy_delay.1, 1, lt\n" \ don't enter loop if input is 0
    C" __busy_delay.0:\n"
        C" sub r0, r0, 1\n" \ 6 cycles
        C" jumpr __busy_delay.0, 0, gt\n" \ 4 cycles, loop if greater than 0
    C" __busy_delay.1:\n"
    C" add r3, r3, 1\n" \ decrement stack
    7 \ 7 strings
;

\ create the token threaded
BUSY_DELAY.BUILDER
\ create the subroutine threaded
BUSY_DELAY.BUILDER
ASSEMBLY-BOTH BUSY_DELAY
TOKEN_NEXT_SKIP_LOAD LAST SET-ULP-ASM-NEXT

: DELAY_MS ( n -- )
    BEGIN
        DUP
    WHILE \ while n is not 0
        1033 BUSY_DELAY \ calibrated at 21 C
        1-
    REPEAT
    DROP \ remove n
;
