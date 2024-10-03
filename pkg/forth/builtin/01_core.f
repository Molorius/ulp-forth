
32 WORD \ --CREATE-FORTH ] 10 WORD DROP EXIT [ \ end of line comments work now
32 WORD BL --CREATE-FORTH ] 32 EXIT [ \ useful to not repeat the ascii for space
BL WORD CREATE --CREATE-FORTH ] BL WORD --CREATE-FORTH EXIT [ \ easier to create definitions
CREATE SEE ] BL WORD --SEE EXIT [ \ not core but really nice for debugging
CREATE ( ] ')' WORD DROP EXIT [ ( inline comments work now )
CREATE TRUE ] -1 EXIT [ ( -- true )
CREATE FALSE ] 0 EXIT [ ( -- false )
