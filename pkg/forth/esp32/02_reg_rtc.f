\ Copyright 2024-2025 Blake Felt blake.w.felt@gmail.com
\ This Source Code Form is subject to the terms of the Mozilla Public
\ License, v. 2.0. If a copy of the MPL was not distributed with this
\ file, You can obtain one at https://mozilla.org/MPL/2.0/.

0x3FF4800C CONSTANT RTC_CNTL_TIME_UPDATE_REG
31 CONSTANT RTC_CNTL_TIME_UPDATE_S
30 CONSTANT RTC_CNTL_TIME_VALID_S

0x3FF48010 CONSTANT RTC_CNTL_TIME0_REG
0x3FF48014 CONSTANT RTC_CNTL_TIME1_REG
0x3FF480C0 CONSTANT RTC_CNTL_LOW_POWER_ST_REG
27 CONSTANT RTC_CNTL_MAIN_STATE_IN_IDLE_S
19 CONSTANT RTC_CNTL_RTC_RDY_FOR_WAKEUP_S
