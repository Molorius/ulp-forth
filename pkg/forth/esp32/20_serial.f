
\ Create the assembly to bitbang serial writes.
\ Takes in a pin number and wait time, then parses
\ the next word. Creates a definition for that word
\ that can be used to write one character over serial.
\
\ This is in assembly because the forth timings
\ may change later with optimization passes, assembly
\ is consistent. It's also difficult to tune high baud
\ rates in forth without taking a lot of space.
: SERIAL.WRITE_CREATE ( pin wait-time )
    2>R \ put arguments on return stack while building
    C" ld r0, r3, 0\n" \ load the character
    C" add r3, r3, 1\n" \ decrement the return stack
    C" or r0, r0, 0x100\n" \ set the end bit
    C" lsh r0, r0, 1\n" \ set the start bit
    C" stage_rst\n" \ reset the stage counter
    C" __serial_write_0_" 1 RPICK C" _" 0 RPICK C" :\n"
        C" and r1, r0, 1\n" \ get the lowest bit
        C" jump __serial_write_1_" 1 RPICK C" _" 0 RPICK C" , eq\n"
            \ set high
            RTCIO_RTC_GPIO_OUT_W1TS_REG
            RTCIO_RTC_GPIO_OUT_DATA_W1TS_S 1 RPICK +
            1 1
            WRITE_RTC_REG.BUILDER >C
            C" jump __serial_write_2_" 1 RPICK C" _" 0 RPICK C" \n"
        C" __serial_write_1_" 1 RPICK C" _" 0 RPICK C" :\n"
            \ set low
            RTCIO_RTC_GPIO_OUT_W1TC_REG
            RTCIO_RTC_GPIO_OUT_DATA_W1TC_S 1 RPICK +
            1 1
            WRITE_RTC_REG.BUILDER >C
            C" jump __serial_write_2_" 1 RPICK C" _" 0 RPICK C" \n"
        C" __serial_write_2_" 1 RPICK C" _" 0 RPICK C" :\n"
        C" wait " R@ C" \n"
        C" rsh r0, r0, 1\n"
        C" stage_inc 1\n"
        C" jumps __serial_write_0_" 1 RPICK C" _" 0 RPICK C" , 10, lt\n"
    C" jump __next_skip_r2"
    47 C> C> + + \ get the total number of inputs
    BL WORD --CREATE-ASSEMBLY \ create the assembly
    2R> 2DROP \ clean up the return stack
;

\ these were found at 21 C with a logic analyzer
\ 0 CONSTANT SERIAL.WRITE_4800_BAUD
874 CONSTANT SERIAL.WRITE_9600_BAUD
\ 0 CONSTANT SERIAL.WRITE_19200_BAUD
\ 0 CONSTANT SERIAL.WRITE_38400_BAUD
\ 0 CONSTANT SERIAL.WRITE_57600_BAUD
34  CONSTANT SERIAL.WRITE_115200_BAUD \ fastest we can go with this algorithm
