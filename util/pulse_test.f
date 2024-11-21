\ Copyright 2024 Blake Felt blake.w.felt@gmail.com
\ This Source Code Form is subject to the terms of the Mozilla Public
\ License, v. 2.0. If a copy of the MPL was not distributed with this
\ file, You can obtain one at https://mozilla.org/MPL/2.0/.

\ Pulses out on a pin, used to calibrate the millisecond delay
\ and to compare the fastest assembly to fastest forth.
\ Should be compiled without optimizations.

\ rtc_gpio14 == gpio13

: pin_init
    rtc_gpio14.enable
    rtc_gpio14.output_enable
    rtc_gpio14.set_low
;

\ create the assembly to toggle a pin as fast as possible
string" reg_wr 257, 28, 28, 1\nreg_wr 258, 28, 28, 1\n"
string" unique_label_to_skip_assembly_optimizations_in_pulse_test_f:\n"
string" jump __next_skip_load"
3 \ 3 assembly objects to compile
assembly asm-pulse \ create the assembly word named "asm-pulse"

: main
    pin_init

    begin
        \ toggle in assembly
        asm-pulse
        \ toggle in forth
        rtc_gpio14.set_high rtc_gpio14.set_low
        \ toggle in forth with 0 delay
        0 rtc_gpio14.set_high delay_ms rtc_gpio14.set_low
        \ small delay
        1 rtc_gpio14.set_high delay_ms rtc_gpio14.set_low
        \ bigger
        10 rtc_gpio14.set_high delay_ms rtc_gpio14.set_low
        \ big!
        1000 rtc_gpio14.set_high delay_ms rtc_gpio14.set_low

        100 delay_ms
    again
;
