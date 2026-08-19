\ control.fth - conditionals, loops, case, recursion

: c1 if 42 else 7 then ;
T{ true c1 -> 42 }T
T{ false c1 -> 7 }T

: c2 if 1 then ;
T{ false c2 -> }T

: nested ( a b -- n ) if if 3 else 2 then else drop 1 then ;
T{ true true nested -> 3 }T
T{ false true nested -> 2 }T
T{ true false nested -> 1 }T

: countdown 0 begin 1+ dup 5 = until ;
T{ countdown -> 5 }T

: whileloop 0 begin dup 5 < while 1+ repeat ;
T{ whileloop -> 5 }T

: sum-to 0 swap 0 do i + loop ;
T{ 5 sum-to -> 10 }T

: step2 0 10 0 do i + 2 +loop ;
T{ step2 -> 20 }T

: down 0 0 5 do 1+ -1 +loop ;
T{ down -> 6 }T

: with-leave 0 100 0 do 1+ dup 3 = if leave then loop ;
T{ with-leave -> 3 }T

: q-do 0 5 5 ?do 1+ loop ;
T{ q-do -> 0 }T
: q-do2 0 5 0 ?do 1+ loop ;
T{ q-do2 -> 5 }T

: multi 0 3 0 do 3 0 do i j + + loop loop ;
T{ multi -> 18 }T

: colours case
    1 of 11 endof
    2 of 22 endof
    99 swap
  endcase ;
T{ 1 colours -> 11 }T
T{ 2 colours -> 22 }T
T{ 5 colours -> 99 }T

: fact dup 1 > if dup 1- recurse * then ;
T{ 5 fact -> 120 }T
T{ 1 fact -> 1 }T

: fib dup 2 < if exit then dup 1- recurse swap 2 - recurse + ;
T{ 10 fib -> 55 }T

: early 1 exit 2 ;
T{ early -> 1 }T

\ interpret-time control flow
T{ 0 5 0 do i + loop -> 10 }T
T{ 0 begin 1+ dup 3 = until -> 3 }T
