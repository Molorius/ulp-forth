

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
STRING" st r0, r3, 0\njump __next_skip_load"
5 C> C> C> C> + + + + \ add up the strings and the built instructions
BL WORD RTC_CLOCK --CREATE-ASSEMBLY \ create RTC_CLOCK

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

STRING" ld r0, r3, 0\n"
STRING" jumpr __busy_delay.1, 1, lt\n" \ don't enter loop if input is 0
STRING" __busy_delay.0:\n"
    STRING" sub r0, r0, 1\n" \ 6 cycles
    STRING" jumpr __busy_delay.0, 0, gt\n" \ 4 cycles, loop if greater than 0
STRING" __busy_delay.1:\n"
STRING" add r3, r3, 1\n" \ decrement stack
STRING" jump __next_skip_load" \ exit
8 BL WORD BUSY_DELAY --CREATE-ASSEMBLY \ create BUSY_DELAY

: DELAY_MS ( n -- )
    BEGIN
        DUP
    WHILE \ while n is not 0
        1033 BUSY_DELAY \ calibrated at 21 C
        1-
    REPEAT
    DROP \ remove n
;
