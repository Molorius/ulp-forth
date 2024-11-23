\ Copyright 2024 Blake Felt blake.w.felt@gmail.com
\ This Source Code Form is subject to the terms of the Mozilla Public
\ License, v. 2.0. If a copy of the MPL was not distributed with this
\ file, You can obtain one at https://mozilla.org/MPL/2.0/.

STRING" __WAKE.0:\n"
    \ check if esp32 is ready for wakeup
    RTC_CNTL_LOW_POWER_ST_REG RTC_CNTL_RTC_RDY_FOR_WAKEUP_S 1
    READ_RTC_REG.BUILDER >C
    \ exit if ready
    STRING" jumpr __WAKE.1, 1, lt\n"
    \ check if esp32 is in a sleep mode
    RTC_CNTL_LOW_POWER_ST_REG RTC_CNTL_MAIN_STATE_IN_IDLE_S 1
    READ_RTC_REG.BUILDER >C
    \ loop if not
    STRING" jumpr __WAKE.0, 0, gt\n"
STRING" __WAKE.1:\n"
\ wake!
STRING" wake\n"
STRING" jump __next_skip_load"
6 C> C> + + \ add up the strings and the built instructions
ASSEMBLY WAKE \ create WAKE
