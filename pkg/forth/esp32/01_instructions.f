\ Copyright 2024-2025 Blake Felt blake.w.felt@gmail.com
\ This Source Code Form is subject to the terms of the Mozilla Public
\ License, v. 2.0. If a copy of the MPL was not distributed with this
\ file, You can obtain one at https://mozilla.org/MPL/2.0/.

\ This file contains words to create custom ULP assembly,
\ specifically meant for hardware access.
\ Note that words created by ASSEMBLY should NOT be used
\ to access the return stack.

0x3ff48000 CONSTANT DR_REG_RTCCNTL_BASE

: RTC_ADDR_FIX ( addr -- addr-fixed )
    2 RSHIFT \ divide by 4
    0x3FF AND \ remove extra bits
;

: REG_RD.BUILDER ( addr high low -- objn .. obj0 n )
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

: READ_RTC_ADDR_CHANGE ( addr low width -- addr-fixed high low )
    >R >R
    RTC_ADDR_FIX
    R> R> ( addr-fixed low width )
    SWAP DUP >R ( addr-fixed width low R: low )
    + 1 - ( addr-fixed high R: low )
    R> ( addr-fixed high low )
;

: READ_RTC_REG.BUILDER ( addr low width "<spaces>name" -- )
    READ_RTC_ADDR_CHANGE
    REG_RD.BUILDER
;

: READ_RTC_REG ( addr low width )
    2 PICK 2 PICK 2 PICK \ duplicate the inputs
    >R >R >R
    \ create the token threaded assembly
    READ_RTC_REG.BUILDER
    C" sub r3, r3, 1\nst r0, r3, 0" \ increase stack, store result
    SWAP 1 + \ number of inputs
    \ get the inputs
    R> R> R>
    \ create the subroutine threaded assembly
    READ_RTC_REG.BUILDER
    C" sub r3, r3, 1\nst r0, r3, 0" \ increase stack, store result
    SWAP 1 + \ number of inputs
    ASSEMBLY-BOTH
    TOKEN_NEXT_SKIP_LOAD LAST SET-ULP-ASM-NEXT
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

: WRITE_RTC_ADDR_CHANGE ( addr low width data -- addr high low data )
    >R \ push data onto return stack
    READ_RTC_ADDR_CHANGE \ fix the rest
    R> \ pop data
;

: WRITE_RTC_REG.BUILDER
    WRITE_RTC_ADDR_CHANGE
    REG_WR.BUILDER
;

\ create an assembly word that writes to an RTC register
: WRITE_RTC_REG ( addr low width data "<spaces>name" -- -- )
    3 PICK 3 PICK 3 PICK 3 PICK \ duplicate the inputs
    >R >R >R >R
    \ create the token threaded assembly
    WRITE_RTC_REG.BUILDER
    \ get the inputs
    R> R> R> R>
    \ create the subroutine threaded assembly
    WRITE_RTC_REG.BUILDER
    ASSEMBLY-BOTH \ create the assembly
    TOKEN_NEXT_SKIP_LOAD LAST SET-ULP-ASM-NEXT
;

\ create an assembly word that writes to two RTC registers
: 2WRITE_RTC_REG ( addr0 high0 low0 data0 addr1 high1 low1 data1 "<spaces>name" -- )
    \ dup the inputs
    7 PICK 7 PICK 7 PICK 7 PICK 7 PICK 7 PICK 7 PICK 7 PICK
    >R >R >R >R >R >R >R >R
    \ create the token threaded assembly
    >R >R >R >R
    WRITE_RTC_REG.BUILDER >C
    R> R> R> R>
    WRITE_RTC_REG.BUILDER C> +
    \ get the inputs
    R> R> R> R> R> R> R> R>
    \ create the subroutine threaded assembly
    >R >R >R >R
    WRITE_RTC_REG.BUILDER >C
    R> R> R> R>
    WRITE_RTC_REG.BUILDER C> +
    ASSEMBLY-BOTH
    TOKEN_NEXT_SKIP_LOAD LAST SET-ULP-ASM-NEXT
;
