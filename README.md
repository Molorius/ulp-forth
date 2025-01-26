# ulp-forth

[![Latest Release](https://img.shields.io/github/v/release/Molorius/ulp-forth?sort=semver)](https://github.com/Molorius/ulp-forth/releases)
[![Test Status](https://github.com/Molorius/ulp-forth/actions/workflows/tests.yml/badge.svg)](https://github.com/Molorius/ulp-forth/actions?query=workflow%3Atests)
[![License: MPL 2.0](https://img.shields.io/badge/License-MPL%202.0-brightgreen.svg)](https://opensource.org/licenses/MPL-2.0)

ulp-forth is a Forth interpreter and optimizing cross compiler for the 
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

Copyright 2024-2025 Blake Felt blake.w.felt@gmail.com

# Contents
* [Installation](#installation)
* [Building ulp-forth](#building-ulp-forth)
* [Using ulp-forth](#using-ulp-forth)
* [Sharing memory](#sharing-memory)
* [Threading models](#threading-models)
* [Assembly words](#assembly-words)
* [System words](#system-words)
* [Clock words](#clock-words)
* [GPIO words](#gpio-words)
* [Serial words](#serial-words)
* [I2C words](#i2c-words)
* [Standard Core words](#standard-core-words)
* [Standard Core Extension words](#standard-core-extension-words)
* [Standard Double words](#standard-double-words)
* [Optimizations](#optimizations)

# Installation

Releases can be found on the [release page](https://github.com/Molorius/ulp-forth/releases).

You can also build the latest tagged version from source with:
```bash
go install github.com/Molorius/ulp-forth@latest
```

# Developing ulp-forth

The compiler can be built from source with
```
go build
```
Unit tests are run on the 
host and on a ulp emulator, they can be run with
```
go test ./...
```

# Using ulp-forth

Help running the program can be found with `ulp-forth --help`.

## Running the interpreter

The interpreter can be run with 
```
ulp-forth run
```
This can be used for testing logic, but only runs on the host so
cannot be used for testing hardware. Type `bye` to exit.

You can load files before the interpreter starts by including them
in run the command.
```
ulp-forth run your_code.f
```

## Running the compiler

The cross compiler can be run with 
```
ulp-forth build your_code.f
```
The user should pass in the list of files to be built, which will be interpreted in order before cross compiling the `MAIN` word. So 
```
ulp-forth build first.f second.f
```
will first interpret first.f, then second.f, then cross compile the `MAIN` word.

## Compiler flags

* `--assembly` Output assembly that can be compiled by the main assemblers, set the --reserved flag before using.
* `--custom_assembly` Output assembly only for use by ulp-asm, another project by this author.

* `--output` Name of the output file.
* `--reserved` Number of reserved bytes for the ULP, for use with --assembly flag (default 8176). Note that the Espressif linker has a bug so has 12 less total bytes. Any space not used by code or data is used for the stacks.
* `--subroutine` Use the subroutine threading model, see the [threading models](#threading-models) section. Faster but larger.


# Sharing memory

There are words that can be used to share memory with the esp32. When compiled with the `--custom_assembly` or `--assembly` flags, the output assembly will include the `.global` directive for the associated memory. This memory will not be optimized away.

| Shared word        | Equivalent word |
| ------------------ | --------------- |
| `GLOBAL-VARIABLE`  | `VARIABLE`      |
| `GLOBAL-2VARIABLE` | `2VARIABLE`     |
| `GLOBAL-ALLOCATE`  | `ALLOCATE`      |

Access should be done while holding the mutex, see the [System words](#system-words) section.

Example:

```
global-variable example \ create a global variable named "example"

\ read an address while holding the mutex
: global@ ( address -- n )
    mutex.take \ take ownership of the mutex
    @ \ read the value at the address
    mutex.give \ release the mutex
;

\ write to an address while holding the mutex
: global! ( n address -- )
    mutex.take \ take ownership of the mutex
    ! \ write the value to the address
    mutex.give \ release the mutex
;

\ get the value at "example"
: get-example ( -- )
    example \ put the address of the memory onto the stack
    global@ \ read it
;

\ set the value at "example"
: set-example ( n -- )
    example \ put the address of the memory onto the stack
    global! \ write to it
;
```

# Threading models

There are two threading models for the output ULP code. This is the forth definition of "threading" and is not the same as multithreading in other languages. It can be thought of as the execution environment.

Token threading is usually smaller and subroutine threading is usually faster, but this can vary based on the program and optimizations.

## Token threading (default)
This uses a lightweight virtual machine to execute all forth words. This allows for some very compact code, but there is a speed penalty for the virtual machine.

Code using this is roughly 20% smaller than subroutine threaded code.

## Subroutine threading

This can be enabled with the `--subroutine` flag. It compiles all forth words into assembly subroutines. This is very fast while executing, but there is a size penalty.

Code using this is roughly 20% faster than token threaded code.

# Assembly words

A few words are provided to make ULP assembly without extending
the compiler.

## `ASSEMBLY`
```
ASSEMBLY ( objn objn-1 ... obj0 n "\<spaces\>name" -- )
```

Skip leading spaces. Parse `name` delimited by a space. Create a
definition for `name` that compiles to token threaded ULP assembly.
The assembly is the contents of the objects on the stack, with object count `n`.
Objects can be strings or integers.

Note that the assembly is built with ulp-asm, a project by the same author
as ulp-forth. It is slightly different than the Espressif or
micropython compilers.

Words built with ASSEMBLY should not access the return stack.

Example:
```
c" move r0, "
0x10 \ include this number
c" \njump next"
3 \ we want to compile the 3 items on the stack
ASSEMBLY MY-EXAMPLE
```

That will create a word `MY-EXAMPLE` which will output the
token threaded assembly:
```assembly
move r0, 16
jump next
```

## `ASSEMBLY-SRT`
```
ASSEMBLY-SRT ( objn objn-1 ... obj0 n "\<spaces\>name" -- )
```

Skip leading spaces. Parse `name` delimited by a space. Create a
definition for `name` that compiles to subroutine threaded ULP assembly.
The assembly is the contents of the objects on the stack, with object count `n`.
Objects can be strings or integers.

Note that the assembly is built with ulp-asm, a project by the same author
as ulp-forth. It is slightly different than the Espressif or
micropython compilers.

Words built with ASSEMBLY-SRT should not access the return stack.
A return is automatically appended.

Example:
```
c" move r0, "
0x10 \ include this number
\ 
2 \ we want to compile the 2 items on the stack
ASSEMBLY-SRT MY-EXAMPLE
```

That will create a word `MY-EXAMPLE` which will output the
subroutine threaded assembly:
```assembly
move r0, 16
add r2, r2, 1
jump r2
```

## `ASSEMBLY-BOTH`
```
ASSEMBLY-BOTH
  (
    objAn objAn-1 ... objA0 n
    objBn objBn-1 ... objB0 m
    "\<spaces\>name" --
  )
```

Skip leading spaces. Parse `name` delimited by a space. Create a
definition for `name` that compiles to token threaded and subroutine
threaded ULP assembly.

The token threaded assembly is the contents of the `objA` objects
on the stack, with object count `n`.
The subroutine threaded assembly is the contents of the `objB` objects
on the stack, with object count `m`.
Objects can be strings or integers.

Note that the assembly is built with ulp-asm, a project by the same author
as ulp-forth. It is slightly different than the Espressif or
micropython compilers.

Words built with ASSEMBLY-BOTH should not access the return stack.
A return is automatically appended to the subroutine threaded assembly.

Example:
```
\ token threaded
c" move r0, "
0x10 \ include this number
2 \ we want to compile the 2 items on the stack

\ subroutine threaded
c" move r1, "
0x11 \ include this number
2 \ we want to compile the 2 items on the stack

ASSEMBLY-BOTH MY-EXAMPLE
```

That will create a word `MY-EXAMPLE` which will output the
token threaded assembly:
```assembly
move r0, 16
jump next
```

and the subroutine threaded assembly:
```assembly
move r1, 17
add r2, r2, 1
jump r2
```

## `READ_RTC_REG`

```
READ_RTC_REG ( addr low width "\<spaces\>name" -- )
```

Skip leading spaces. Parse `name` delimited by a space. Create an
assembly definition for `name` that reads from RTC address `addr`,
low bit `low`, width `width`; the result is placed on the stack.

## `WRITE_RTC_REG`

```
WRITE_RTC_REG ( addr low width data "\<spaces\>name" -- )
```

Skip leading spaces. Parse `name` delimited by a space. Create an
assembly definition for `name` that writes to RTC address `addr`,
low bit `low`, width `width`.

## `2WRITE_RTC_REG`
```
2WRITE_RTC_REG
( addr0 high0 low0 data0
  addr1 high1 low1 data1 "\<spaces\>name" -- )
```

Skip leading spaces. Parse `name` delimited by a space. Create an
assembly definition for `name` that writes to RTC address `addr0`
followed by address `addr1`.

# System words

System words only run on the ULP.

## `HALT`
```
HALT ( -- )
```

Halt execution of the ULP. Execution will resume at the instruction immediately following the HALT on both token and subroutine threaded models.

## `MUTEX.TAKE`
```
MUTEX.TAKE ( -- )
```

Takes the software mutex. The example project includes esp32 code to use this but a better way to use it needs to be written.

## `MUTEX.GIVE`
```
MUTEX.GIVE ( -- )
```

Gives the software mutex. The example project includes esp32 code to use this but a better way to use it needs to be written.

# Clock words

Clock words only run on the ULP.

The ULP has access to the RTC_SLOW clock. ulp-forth also has
some busy-wait words for delays. Note that the ULP runs on the
RTC_FAST clock, which we cannot read.

## `RTC_CLOCK`
```
RTC_CLOCK ( -- d )
```

Read the lower 32 bits of the rtc_slow clock.

## `RTC_CLOCK_DELAY`
```
RTC_CLOCK_DELAY ( d -- )
```

Delay for d rtc_slow ticks.

## `BUSY_DELAY`
```
BUSY_DELAY ( n -- )
```

Delay n times in a tight assembly loop.

## `DELAY_MS`
```
DELAY_MS ( n -- )
```

Delay n milliseconds. The accuracy of this is affected by temperature
and is device dependent.

# GPIO words

The ULP can access certain pins called RTC_GPIO. These are mapped to the GPIO
as well. Words are defined to help interface with them.

Each of these words are written with the prefix GPIOn, where n is the pin number.
There are words for both the RTC_GPIOn and GPIOn numbers. Words are only written
for GPIO that the ULP can access, and if a pin doesn't support output then the output
words aren't defined.

Below is a table of all pins accessible to the ULP. RTC_GPIO is the
naming used by the RTC subsystem, GPIO is the naming used by the
rest of the ESP32 documentation.

| GPIO | RTC_GPIO | Notes |
|------|----------|-------|
| 36   | 0        | Input only, no pullups or pulldowns.
| 37   | 1        | Input only, no pullups or pulldowns.
| 38   | 2        | Input only, no pullups or pulldowns.
| 39   | 3        | Input only, no pullups or pulldowns.
| 34   | 4        | Input only, no pullups or pulldowns.
| 35   | 5        | Input only, no pullups or pulldowns.
| 25   | 6        |
| 26   | 7        |
| 33   | 8        |
| 32   | 9        |
| 4    | 10       |
| 0    | 11       |
| 2    | 12       |
| 15   | 13       |
| 13   | 14       |
| 12   | 15       |
| 14   | 16       |
| 27   | 17       |

Not all pins are tested.

## `GPIOn.ENABLE`
```
GPIOn.ENABLE ( -- )
```

Enable the usage of this pin by the ULP.

## `GPIOn.OUTPUT_ENABLE`
```
GPIOn.OUTPUT_ENABLE ( -- )
```

Enable output on this pin.

## `GPIOn.OUTPUT_DISABLE`
```
GPIOn.OUTPUT_DISABLE ( -- )
```

Disable output on this pin.

## `GPIOn.INPUT_ENABLE`
```
GPIOn.INPUT_ENABLE ( -- )
```

Allow reading from this pin.

## `GPIOn.SET_HIGH`
```
GPIOn.SET_HIGH ( -- )
```

Set this pin to high.

## `GPIOn.SET_LOW`
```
GPIOn.SET_LOW ( -- )
```

Set this pin to low.

## `GPIOn.SET`
```
GPIOn.SET ( n -- )
```

If n is 0, set this pin to low. Otherwise set to high.

## `GPIOn.GET`
```
GPIOn.GET ( -- n )
```

Get the value of this pin. 0 or 1.

## `GPIOn.PULLUP_ENABLE`
```
GPIOn.PULLUP_ENABLE ( -- )
```

Enable the pullup resistors on this pin.

## `GPIOn.PULLUP_DISABLE`
```
GPIOn.PULLUP_DISABLE ( -- )
```

Disable the pullup resistors on this pin.

## `GPIOn.PULLDOWN_ENABLE`
```
GPIOn.PULLDOWN_ENABLE ( -- )
```

Enable the pulldown resistors on this pin.

## `GPIOn.PULLDOWN_DISABLE`
```
GPIOn.PULLDOWN_DISABLE ( -- )
```

Disable the pulldown resistors on this pin.

## `GPIO_NUMBER_TO_RTC`
```
GPIO_NUMBER_TO_RTC ( gpio_num -- rtc_gpio_num )
```

Convert the GPIO number to the corresponding
RTC_GPIO number.

# Serial words

The ULP does not have a hardware serial. This is implemented
in assembly.

## `SERIAL.WRITE_CREATE`
```
SERIAL.WRITE_CREATE ( pin wait-time "\<spaces\>name" -- )
```

Skip leading spaces. Parse `name` delimited by a space. Create an
assembly definition for `name` that uses the `pin` RTC_GPIO and
delay `wait-time` to write to serial. See the "hello_world.f" example for setup.

For example, if you used the word `SERIAL_TX` then it would create the definition:
```
SERIAL_TX ( c -- )
```
which outputs the character `c`.

## `SERIAL.WRITE_9600_BAUD`
```
SERIAL.WRITE_9600_BAUD ( -- wait-time )
```

Returns the wait-time to achieve 9600 baud.

## `SERIAL.WRITE_115200_BAUD`
```
SERIAL.WRITE_115200_BAUD ( -- wait-time )
```

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

See the "util/i2c_scan.f" file for pin setup.

## `I2C.START`
```
I2C.START ( -- )
```

Send a start condition on the bus.

## `I2C.START_READ`
```
I2C.START_READ ( address -- ack )
```

Send the start condition and send a read command to the address.
Returns `TRUE` if acknowledged.

## `I2C.START_WRITE`
```
I2C.START_WRITE ( address -- ack )
```

Send the start condition and send a write command to the address.
Returns `TRUE` if acknowledged.

## `I2C.WRITE`
```
I2C.WRITE ( n -- ack )
```

Send byte `n` on the bus. Returns `TRUE` if acknowledged.

## `I2C.READ`
```
I2C.READ ( -- n )
```

Read a byte `n` from the bus. This does not respond with a ack/nack,
that should be done with `I2C.ACK` or `I2C.NACK`.

## `I2C.ACK`
```
I2C.ACK ( -- )
```

Send an ack bit on the bus.

## `I2C.NACK`
```
I2C.NACK ( -- )
```

Send a nack bit on the bus.

## `I2C.STOP`
```
I2C.STOP ( -- )
```

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
* `,`
  * Can only run on host.
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
* `>BODY`
* `>R`
* `?DUP`
* `@`
* `ABS`
* `ALIGN`
  * Can only run on host.
* `ALIGNED`
* `ALLOT`
  * Can only run on host.
* `AND`
* `BASE`
* `BEGIN`
* `BL`
* `C!`
* `C,`
  * Can only run on host.
* `C@`
* `CELL+`
* `CELLS`
* `CHAR`
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
* `DOES>`
  * Can only run on host.
* `DROP`
* `DUP`
* `ELSE`
* `EMIT`
  * Deferred so a program can output to any interface.
    `EMIT` is used for all printing, such as `.`.
* `EVALUATE`
  * Can only run on host.
* `EXECUTE`
* `EXIT`
* `FILL`
* `FIND`
  * Can only run on host.
* `HERE`
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
* `MOVE`
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
* `SPACES`
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

* `.(`
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
  * (nonstandard) can be interpreted.
* `CASE`
* `COMPILE,`
* `DEFER`
* `DEFER!`
* `DEFER@`
* `ENDCASE`
* `ENDOF`
* `ERASE`
* `FALSE`
* `HEX`
* `IS`
* `NIP`
* `OF`
* `PICK`
* `ROLL`
* `TO`
* `TRUE`
* `TUCK`
* `U>`
* `VALUE`
* `WITHIN`
* `[COMPILE]`
* `\`

# Standard Double words

Not all of the double words are currently implemented, but they
all can be in the future.

* `2CONSTANT`
* `2LITERAL`
* `2VARIABLE`
* `D+`
* `D-`
* `D0<`
* `D<`
* `D0=`
* `D>S`
* `DABS`
* `DMAX`
* `DMIN`
* `DNEGATE`
* `M+`

# Standard Double extension words

* `2VALUE`
* `DU<`

# Optimizations

The cross compiler includes some optimizations. More may be added later.

* Inline deferred words. If a word is defined by `DEFER` but the deferred word cannot be changed by the cross compiled output, this will inline the word. Normally a deferred word will look like: `address-containing-word @ EXECUTE EXIT`, this optimizes that to: `word EXIT`.
* Tail calls. Words may be defined in assembly or forth. If the final word before an `EXIT` (or end of definition) is a forth word, this will instead jump to it. For example, a word that is compiled as `+ forth-word EXIT` will be optimized to `+ jump(forth-word)`. Smaller in token threaded model, faster, saves a stack slot.

To be added later:
* Fallthrough forth words instead of tail call
* Assembly inlining (subroutine threaded only)
* Forth inlining
* Forth common sequence compression
* Flow control analysis
* Constant folding
* Peephole optimization
