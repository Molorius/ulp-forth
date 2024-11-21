\ Copyright 2024 Blake Felt blake.w.felt@gmail.com
\ This Source Code Form is subject to the terms of the Mozilla Public
\ License, v. 2.0. If a copy of the MPL was not distributed with this
\ file, You can obtain one at https://mozilla.org/MPL/2.0/.

\ Initializes the led pin. Change to whichever gpio you
\ want to use.
\ You can save space by initializing this with the esp32.
: pin_init ( -- )
    gpio2.enable
    gpio2.output_enable
    \ or use the rtc_gpio equivalents:
    \ rtc_gpio12.enable
    \ rtc_gpio12.output_enable
;

\ Sets the led gpio to high or low.
\ Takes in a value. Set low if 0 else set high.
\ Change to whichever gpio you want to use.
: led_set ( val -- )
    gpio2.set
    \ or use the rtc_gpio equivalent:
    \ rtc_gpio12.set
;

\ The number of milliseconds we want to delay.
1000 constant ms

: main
    pin_init
    begin
        1 led_set \ set high
        \ gpio2.set_high \ or set high directly
        ms delay_ms

        0 led_set \ set low
        \ gpio2.set_low \ or set low directly
        ms delay_ms
    again
;
