\ Copyright 2024-2025 Blake Felt blake.w.felt@gmail.com
\ This Source Code Form is subject to the terms of the Mozilla Public
\ License, v. 2.0. If a copy of the MPL was not distributed with this
\ file, You can obtain one at https://mozilla.org/MPL/2.0/.

: WAKE.BUILDER
    C" __WAKE.0:\n"
        \ check if esp32 is ready for wakeup
        RTC_CNTL_LOW_POWER_ST_REG RTC_CNTL_RTC_RDY_FOR_WAKEUP_S 1
        READ_RTC_REG.BUILDER >C
        \ exit if ready
        C" jumpr __WAKE.1, 1, lt\n"
        \ check if esp32 is in a sleep mode
        RTC_CNTL_LOW_POWER_ST_REG RTC_CNTL_MAIN_STATE_IN_IDLE_S 1
        READ_RTC_REG.BUILDER >C
        \ loop if not
        C" jumpr __WAKE.0, 0, gt\n"
    C" __WAKE.1:\n"
    \ wake!
    C" wake\n"
    5 C> C> + + \ add up the strings and the built instructions
;

\ token threaded assembly
WAKE.BUILDER
\ subroutine threaded assembly
WAKE.BUILDER
ASSEMBLY-BOTH WAKE
TOKEN_NEXT_SKIP_LOAD LAST SET-ULP-ASM-NEXT
