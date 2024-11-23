\ Copyright 2024 Blake Felt blake.w.felt@gmail.com
\ This Source Code Form is subject to the terms of the Mozilla Public
\ License, v. 2.0. If a copy of the MPL was not distributed with this
\ file, You can obtain one at https://mozilla.org/MPL/2.0/.

\ Because we use an instruction pointer to execute instructions,
\ halting will continue execution immediately after the HALT 
\ instruction.
\ Placed here so we can use it in all code after MAIN runs.
STRING" halt"
1 ASSEMBLY HALT \ create HALT
