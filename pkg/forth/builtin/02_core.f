
\ ['] parses the next name and compiles the execution token of that name. Immediate.
: ['] ( -- ) ' POSTPONE LITERAL ; IMMEDIATE

\ RECURSE compiles the most recently defined name. Immediate.
: RECURSE ( -- ) LAST COMPILE, ; IMMEDIATE

