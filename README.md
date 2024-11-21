# ulp-forth

[![License: MPL 2.0](https://img.shields.io/badge/License-MPL%202.0-brightgreen.svg)](https://opensource.org/licenses/MPL-2.0)

ulp-forth is a Forth interpreter and cross compiler for the 
ESP32 ULP coprocessor. It has most of the Forth 2020 standard 
implemented. Code is interpreted on the host machine, when
that is finished it compiles the `main` word for execution by the
ULP.

This is a 16 bit Forth implementation. It is case insensitive.

Several design choices were made to save ULP memory:
* The ULP cannot modify the dictionary or allocate memory.
* Only words used by `main` are cross compiled (pruning).
* There are no error checks in the cross compiled output.

The Forth 2020 standard was followed for this implementation.
There are some missing words, but all of the implemented words
follow the standard and pass the test suite.

Speed of ulp-forth versus assembly is hard to compare because 1) it
depends on how you implement subroutines in assembly, 2) the
algorithms are different on a forth stack vs assembly registers, and
3) many forth words are implemented in assembly optimized for the
ulp-forth calling convention, which would be difficult to implement in
assembly. Toggling a pin is 4 times slower.

Copyright 2024 Blake Felt blake.w.felt@gmail.com

# Contents
* [Building ulp-forth](#building-ulp-forth)
* [Assembly words](#assembly-words)
* [Clock words](#clock-words)
* [GPIO words](#gpio-words)
* [Serial words](#serial-words)
* [I2C words](#i2c-words)
* [Standard Core words](#standard-core-words)
* [Standard Core Extension words](#standard-core-extension-words)
* [Standard Double words](#standard-double-words)

# Building ulp-forth

The compiler can be built with `go build`. Unit tests are run on the 
host and on a ulp emulator, they can be run with `go test ./...`.

# Assembly words

A few words are provided to make ULP assembly without extending
the compiler.

## ASSEMBLY

ASSEMBLY ( objn objn-1 ... obj0 n "\<spaces\>name" -- )

Skip leading spaces. Parse `name` delimited by a space. Create a
definition for `name` that compiles to ULP assembly. The assembly is
the contents of the objects on the stack, with object count `n`.
Objects can be strings or integers.

Note that the assembly is built with ulp-asm, a project by the same author
as ulp-forth. It is slightly different than the Espressif or
micropython compilers.

Words built with ASSEMBLY should not access the return stack.

Example:
```forth
STRING" move r0, "
0x10
STRING" \njump next"
3 \ we want to compile the 3 items on the stack
ASSEMBLY MY-EXAMPLE
```

That will create a word `MY-EXAMPLE` which will output the assembly:
```assembly
move r0, 16
jump next
```

## READ_RTC_REG

READ_RTC_REG ( addr low width "\<spaces\>name" -- )

Skip leading spaces. Parse `name` delimited by a space. Create an
assembly definition for `name` that reads from RTC address `addr`,
low bit `low`, width `width`; the result is placed on the stack.

## WRITE_RTC_REG

WRITE_RTC_REG ( addr low width data "\<spaces\>name" -- )

Skip leading spaces. Parse `name` delimited by a space. Create an
assembly definition for `name` that writes to RTC address `addr`,
low bit `low`, width `width`.

## 2WRITE_RTC_REG

2WRITE_RTC_REG
( addr0 high0 low0 data0
  addr1 high1 low1 data1 "\<spaces\>name" -- )

Skip leading spaces. Parse `name` delimited by a space. Create an
assembly definition for `name` that writes to RTC address `addr0`
followed by address `addr1`.

# Clock words

The ULP has access to the RTC_SLOW clock. ulp-forth also has
some busy-wait words for delays. Note that the ULP runs on the
RTC_FAST clock, which we cannot read.

## RTC_CLOCK

RTC_CLOCK ( -- d )

Read the lower 32 bits of the rtc_slow clock.

## RTC_CLOCK_DELAY

RTC_CLOCK_DELAY ( d -- )

Delay for d rtc_slow ticks.

## BUSY_DELAY

BUSY_DELAY ( n -- )

Delay n times in a tight assembly loop.

## DELAY_MS

DELAY_MS ( n -- )

Delay n milliseconds. The accuracy of this is affected by temperature.

# GPIO words

The ULP can access certain pins called RTC_GPIO. These are mapped to the GPIO
as well. Words are defined to help interface with them.

Each of these words are written with the prefix GPIOn, where n is the pin number.
There are words for both the RTC_GPIOn and GPIOn numbers. Words are only written
for GPIO that the ULP can access, and if a pin doesn't support output then the output
words aren't defined.

Not all pins are tested.

## GPIOn.ENABLE

GPIOn.ENABLE ( -- )

Enable the usage of this pin by the ULP.

## GPIOn.OUTPUT_ENABLE

GPIOn.OUTPUT_ENABLE ( -- )

Enable output on this pin.

## GPIOn.OUTPUT_DISABLE

GPIOn.OUTPUT_DISABLE ( -- )

Disable output on this pin.

## GPIOn.INPUT_ENABLE

GPIOn.INPUT_ENABLE ( -- )

Allow reading from this pin.

## GPIOn.SET_HIGH

GPIOn.SET_HIGH ( -- )

Set this pin to high.

## GPIOn.SET_LOW

GPIOn.SET_LOW ( -- )

Set this pin to low.

## GPIOn.SET

GPIOn.SET ( n -- )

If n is 0, set this pin to low. Otherwise set to high.

## GPIOn.GET

GPIOn.GET ( -- n )

Get the value of this pin. 0 or 1.

## GPIOn.PULLUP_ENABLE

GPIOn.PULLUP_ENABLE ( -- )

Enable the pullup resistors on this pin.

## GPIOn.PULLUP_DISABLE

GPIOn.PULLUP_DISABLE ( -- )

Disable the pullup resistors on this pin.

## GPIOn.PULLDOWN_ENABLE

GPIOn.PULLDOWN_ENABLE ( -- )

Enable the pulldown resistors on this pin.

## GPIOn.PULLDOWN_DISABLE

GPIOn.PULLDOWN_DISABLE ( -- )

Disable the pulldown resistors on this pin.

## GPIO_NUMBER_TO_RTC

GPIO_NUMBER_TO_RTC ( gpio_num -- rtc_gpio_num )

Convert the GPIO number to the corresponding
RTC_GPIO number.

# Serial words

The ULP does not have a hardware serial. This is implemented
in assembly.

## SERIAL.WRITE_CREATE

SERIAL.WRITE_CREATE ( pin wait-time "\<spaces\>name" -- )

Skip leading spaces. Parse `name` delimited by a space. Create an
assembly definition for `name` that uses the `pin` RTC_GPIO and
delay `wait-time` to write to serial. See the "hello_world.f" example for setup.

For example, if you used the word `SERIAL_TX` then it would create the definition
SERIAL_TX ( c -- ) which outputs c.

## SERIAL.WRITE_9600_BAUD

SERIAL.WRITE_9600_BAUD ( -- wait-time )

Returns the wait-time to achieve 9600 baud.

## SERIAL.WRITE_115200_BAUD

SERIAL.WRITE_115200_BAUD ( -- wait-time )

Returns the wait-time to achieve 115200 baud. This is the fastest
standard baud rate available with the assembly algorithm used.

# I2C words

The ULP has hardware I2C but it is difficult to use and limited in its features.
This is a forth software implementation. It has clock stretching and allows for an
arbitrary number of devices, but is slower than hardware.

To use this, you need to implement the deferred words:
* `I2C.SDA_HIGH`
* `I2C.SDA_LOW`
* `I2C.SDA_GET`
* `I2C.SCL_HIGH`
* `I2C.SCL_LOW`
* `I2C.SCL_GET`

See the "i2c_scan.f" util for pin setup.

## I2C.START

I2C.START ( -- )

Send a start condition on the bus.

## I2C.START_READ

I2C.START_READ ( address -- ack )

Send the start condition and send a read command to the address.
Returns TRUE if acknowledged.

## I2C.START_WRITE

I2C.START_READ ( address -- ack )

Send the start condition and send a write command to the address.
Returns TRUE if acknowledged.

## I2C.WRITE

I2C.WRITE ( n -- ack )

Send byte n on the bus. Returns TRUE if acknowledged.

## I2C.READ

I2C.READ ( -- n )

Read a byte from the bus. This does not respond with a ack/nack,
that should be done with I2C.ACK or I2C.NACK.

## I2C.ACK

I2C.ACK ( -- )

Send an ack bit.

## I2C.NACK

I2C.NACK ( -- )

Send a nack bit.

## I2C.STOP

I2C.STOP ( -- )

Send a stop condition on the bus.

# Standard Core words

These are the the core words that are implemented. Missing words
may be implemented in the future, only a few such as `HERE` cannot
be implemented because of the ulp-forth architecture.

Some of the core words are created with `DEFER` so they can be
easily overwritten. These are noted below.

Words that can only run on the host are noted as well. Some words,
such as `DO`, cannot be directly executed on the ULP but words 
created with them can run on the ULP. These are not noted as it
depends on how you attempt to use them, ulp-forth will throw an
error if it cannot be cannot be cross compiled.

* `'`
* `(`
* `*`
* `+`
* `+!`
* `+LOOP`
* `-`
* `.`
* `."`
* `/`
* `/MOD`
  * Deferred to `S/REM`, can defer to `F/MOD`.
* `0<`
* `0=`
* `1+`
* `1-`
* `2!`
* `2*`
* `2!`
* `2/`
* `2@`
* `2DROP`
* `2DUP`
* `2OVER`
* `2SWAP`
* `:`
  * Can only run on host.
* `;`
  * Can only run on host.
* `<`
* `=`
* `>`
* `>R`
* `?DUP`
* `@`
* `ABS`
* `ALIGNED`
* `AND`
* `BASE`
* `BEGIN`
* `BL`
* `C!`
* `C@`
* `CELL+`
* `CELLS`
* `CHAR+`
* `CHARS`
* `CONSTANT`
  * Can only run on host.
* `COUNT`
* `CR`
* `CREATE`
  * Can only run on host.
* `DECIMAL`
* `DEPTH`
* `DO`
* `DROP`
* `DUP`
* `ELSE`
* `EMIT`
  * Deferred so a program can output to any interface.
    `EMIT` is used for all printing, such as `.`.
* `EXECUTE`
* `EXIT`
* `FIND`
  * Can only run on host.
* `I`
* `IF`
* `IMMEDIATE`
* `INVERT`
* `J`
* `LEAVE`
* `LITERAL`
* `LOOP`
* `LSHIFT`
* `MAX`
* `MIN`
* `MOD`
* `NEGATE`
* `OR`
* `OVER`
* `POSTPONE`
* `R>`
* `R@`
* `RECURSE`
* `REPEAT`
* `ROT`
* `RSHIFT`
* `S"`
* `S>D`
* `SPACE`
* `STATE`
* `SWAP`
* `THEN`
* `TYPE`
* `U.`
* `U<`
* `UNLOOP`
* `UNTIL`
* `VARIABLE`
* `WHILE`
* `WORD`
  * Can only run on host.
* `XOR`
* `[`
* `[']`
* `[CHAR]`
* `]`

# Standard Core extension words
Missing words may be implemented in the future.

* `0<>`
* `0>`
* `2>R`
* `2R>`
* `2R@`
* `:NONAME`
  * Can only run on host.
* `<>`
* `?DO`
* `AGAIN`
* `BUFFER:`
* `C"`
* `CASE`
* `COMPILE,`
* `DEFER`
* `DEFER!`
* `DEFER@`
* `ENDCASE`
* `ENDOF`
* `FALSE`
* `HEX`
* `IS`
* `NIP`
* `OF`
* `PICK`
* `ROLL`
* `TRUE`
* `TUCK`
* `U>`
* `WITHIN`
* `[COMPILE]`
* `\`

# Standard Double words

Not all of the double words are currently implemented, but they
all can be in the future.

* `2CONSTANT`
* `2LITERAL`
* `2VARIABLE`
* `D-`
* `D0<`
* `D0=`
* `D>S`
* `DABS`
* `DMAX`
* `DMIN`
* `DNEGATE`
* `DU<`
