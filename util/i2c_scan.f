\ Copyright 2024-2025 Blake Felt blake.w.felt@gmail.com
\ This Source Code Form is subject to the terms of the Mozilla Public
\ License, v. 2.0. If a copy of the MPL was not distributed with this
\ file, You can obtain one at https://mozilla.org/MPL/2.0/.

\ This uses a bit-banged i2c implementation to
\ scan for all devices on the bus.
\ The bus should be pulled up with a resistor,
\ 4K7 Ohm is a good value for general use.

\ Initializes the pins.
: pin_init
    \ serial pin
    gpio13.enable
    gpio13.output_enable
    gpio13.set_high
    \ sda pin
    gpio26.enable
    gpio26.input_enable
    gpio26.set_low
    \ scl pin
    gpio27.enable
    gpio27.input_enable
    gpio27.set_low
;

\ uses gpio 13 to serial write at the desired baud.
13 gpio_number_to_rtc serial.write_9600_baud serial.write_create tx
' tx is emit

\ sda pin declarations
' gpio26.output_disable is i2c.sda_high
' gpio26.output_enable is i2c.sda_low
' gpio26.get is i2c.sda_get

\ scl pin declarations
' gpio27.output_disable is i2c.scl_high
' gpio27.output_enable is i2c.scl_low
' gpio27.get is i2c.scl_get

: main
    pin_init

    ." Scanning..." cr
    0x78 0 do \ for every i2c address
        i i2c.start_write \ send a write declaration to the address, ack is on stack
        i2c.stop \ send a stop condition
        if \ if there was an ack
            ." Device found at 0x"
            i u. \ print the number
            cr
        then
    loop
    ." Done." cr
;

hex \ print results in hexadecimal
