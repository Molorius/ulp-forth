\ Copyright 2024 Blake Felt blake.w.felt@gmail.com
\ This Source Code Form is subject to the terms of the Mozilla Public
\ License, v. 2.0. If a copy of the MPL was not distributed with this
\ file, You can obtain one at https://mozilla.org/MPL/2.0/.

\ Initializes the serial out pin.
\ This can be done on the esp32 to save ulp space.
: pin_init
    gpio13.enable
    gpio13.output_enable
    gpio13.set_high
;

\ create a new word called "tx" that
\ uses gpio 13 to serial write at the desired baud.
13 gpio_number_to_rtc serial.write_9600_baud serial.write_create tx
\ 13 gpio_number_to_rtc serial.write_115200_baud serial.write_create tx

\ use our tx word for printing
' tx is emit

\ The number of milliseconds between loops.
1000 constant ms

: main
    pin_init

    0 \ create a loop counter
    begin
        ." Hello world! " \ print the string
        dup u. \ print the loop counter, unsigned
        1 + \ increase the counter
        cr \ print carriage return
        ms delay_ms
    again
;
