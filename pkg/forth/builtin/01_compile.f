
32 WORD IMMEDIATE --CREATE-FORTH ] -1 LAST SET-IMMEDIATE EXIT [
32 WORD \ --CREATE-FORTH ] 10 WORD DROP EXIT [ IMMEDIATE \ end of line comments work now
32 WORD ( --CREATE-FORTH ] ')' WORD DROP EXIT [ IMMEDIATE ( inline comments work now )

\ This file contains words related to compilation. More specifically, all of the 
\ brackets and parantheses confuse my editor so the offending words (and their dependencies)
\ are put in this file.

\ BL places the character for a space on the stack.
32 WORD BL ( -- char ) --CREATE-FORTH ] 32 EXIT [

\ CREATE parses the next name and creates an empty Forth definition
\ for that name.
BL WORD CREATE ( -- ) --CREATE-FORTH ] BL WORD --CREATE-FORTH EXIT [

\ SEE parses the next name and prints the definition of that word.
\ An error is thrown if there is not a word with that name in the dictionary.
CREATE SEE ] BL WORD --SEE EXIT [ \ not core but really nice for debugging

\ TRUE places the 'true' value onto the stack.
CREATE TRUE ( -- true ) ] -1 EXIT [

\ FALSE places the 'false' value onto the stack.
CREATE FALSE ( -- false ) ] 0 EXIT [

\ ' (tick) parses the next name and places the execution token of that name
\ onto the stack.
CREATE ' ] BL WORD --' EXIT [

\ POSTPONE parses the next name and compiles the compilation semantics of that word
\ onto the latest word. Immediate.
CREATE POSTPONE ] ' --POSTPONE EXIT [ IMMEDIATE

\ : parses the next name, creates a dictionary entry for it, hides that word,
\ and puts the VM into compile state. 
\ Used to start compilation.
CREATE : ( -- ) ]
    CREATE \ create the new dictionary entry
    TRUE LAST SET-HIDDEN \ hide it
    ] \ and put in compile mode
EXIT [

\ ; appends EXIT to the most recent definition, unhides it, and puts the VM into interpret state.
\ Used to end compilation. Immediate.
CREATE ; ( -- ) ]
    POSTPONE EXIT \ compile EXIT into new word
    FALSE LAST SET-HIDDEN \ unhide the new word
    POSTPONE [ \ put back in interpret mode
EXIT [ IMMEDIATE
