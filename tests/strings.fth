\ strings.fth - strings, characters and number formatting

T{ s" abc" nip -> 3 }T
T{ s" abc" drop c@ -> 97 }T
T{ char A -> 65 }T
: bracket-char [char] B ;
T{ bracket-char -> 66 }T
T{ bl -> 32 }T
T{ c" hey" count nip -> 3 }T
T{ s" ab" s" ab" str= -> true }T
T{ s" ab" s" ac" str= -> false }T
T{ s" ab" s" abc" str= -> false }T
T{ s" " s" " str= -> true }T

create sbuf 32 allot
T{ s" hello" sbuf swap move sbuf 5 s" hello" str= -> true }T
T{ sbuf 5 char x fill sbuf c@ -> 120 }T
T{ sbuf 5 erase sbuf c@ -> 0 }T

\ numeric conversion
T{ s" 1234" number -> 1234 true }T
T{ s" 12x4" number -> 0 false }T
T{ s" -99" number -> -99 true }T

\ base handling
T{ base @ -> 10 }T
T{ hex base @ decimal -> 16 }T
T{ $ff -> 255 }T
T{ %1010 -> 10 }T
T{ #42 -> 42 }T

\ evaluate
T{ s" 3 4 +" evaluate -> 7 }T

\ pad is usable scratch space
T{ 65 pad c! pad c@ -> 65 }T
