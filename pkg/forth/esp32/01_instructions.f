
\ This file contains words to create custom ULP assembly,
\ specifically meant for hardware access.
\ Note that words created by --CREATE-ASSEMBLY should NOT be used
\ to access the return stack.

0x3ff48000 CONSTANT DR_REG_RTCCNTL_BASE

: RTC_ADDR_FIX
    DR_REG_RTCCNTL_BASE - \ subtract off the offset
    2 RSHIFT \ divide by 4
    0x3FF AND \ remove extra bits
;

: REG_RD.BUILDER
    >R >R >R ( R: low high addr )
    C" reg_rd "
    R> \ addr
    C" , "
    R> \ high
    C" , "
    R> \ low
    C" \n"
    7 \ number of inputs
;

: REG_RD ( addr high low -- )
    REG_RD.BUILDER
    C" sub r3, r3, 1\nst r0, r3, 0\njump __next_skip_load" \ increase stack, store result, next
    SWAP 1 + \ number of inputs
    BL WORD --CREATE-ASSEMBLY
;

: READ_RTC_ADDR_CHANGE ( addr low width -- addr high low )
    >R >R
    RTC_ADDR_FIX
    R> R>
    SWAP DUP >R
    + 1- R>
;

: READ_RTC_REG.BUILDER
    READ_RTC_ADDR_CHANGE
    REG_RD.BUILDER
;

: READ_RTC_REG ( addr low width )
    READ_RTC_REG.BUILDER
    C" sub r3, r3, 1\nst r0, r3, 0\njump __next_skip_load" \ increase stack, store result, next
    SWAP 1 + \ number of inputs
    BL WORD --CREATE-ASSEMBLY
;

: REG_WR.BUILDER ( addr high low data -- strn..str0 n )
    >R >R >R >R
    C" reg_wr "
    R>
    C" , "
    R>
    C" , "
    R>
    C" , "
    R>
    C" \n"
    9 \ number of inputs
;

\ create an assembly word that writes to an RTC register
: REG_WR ( addr high low data )
    REG_WR.BUILDER
    C" jump __next_skip_load"
    SWAP 1 +
    BL WORD --CREATE-ASSEMBLY
;

\ create an assembly word that writes to two RTC registers
: 2REG_WR ( addr0 high0 low0 data0 addr1 high1 low1 data1 -- )
    >R >R >R >R
    REG_WR.BUILDER >C \ put on control flow stack for now
    R> R> R> R>
    REG_WR.BUILDER C> +
    C" jump __next_skip_load"
    SWAP 1 +
    BL WORD --CREATE-ASSEMBLY
;

: WRITE_RTC_ADDR_CHANGE
    >R >R >R
    RTC_ADDR_FIX
    R> R> R>
    >R SWAP DUP >R
    + 1- R> R>
;

: WRITE_RTC_REG.BUILDER
    WRITE_RTC_ADDR_CHANGE
    REG_WR.BUILDER
;

\ create an assembly word that writes to an RTC register
: WRITE_RTC_REG ( addr low width data -- )
    WRITE_RTC_REG.BUILDER
    C" jump __next_skip_load"
    SWAP 1 +
    BL WORD --CREATE-ASSEMBLY
;

\ create an assembly word that writes to two RTC registers
: 2WRITE_RTC_REG ( addr0 high0 low0 data0 addr1 high1 low1 data1 -- )
    >R >R >R >R
    WRITE_RTC_REG.BUILDER >C
    R> R> R> R>
    WRITE_RTC_REG.BUILDER C> +
    C" jump __next_skip_load"
    SWAP 1 +
    BL WORD --CREATE-ASSEMBLY
;
