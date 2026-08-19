\ kernel.fth - the Forth-level part of the system, interpreted at startup.
\ Anything expressible in Forth lives here rather than in Go.

\ --- constants ---
7 constant bel
8 constant bs
9 constant tab-char
10 constant nl
27 constant esc

\ --- stack and variable shorthands ---
: 3dup ( a b c -- a b c a b c ) dup 2over rot ;
: cell- ( a -- a' ) cell - ;
: on ( addr -- ) -1 swap ! ;
: off ( addr -- ) 0 swap ! ;
: incr ( addr -- ) 1 swap +! ;
: decr ( addr -- ) -1 swap +! ;
: ? ( addr -- ) @ . ;
: sq ( n -- n*n ) dup * ;
: 0max ( n -- n ) 0 max ;

\ --- ranges ---
: 0<= ( n -- f ) 0 <= ;
: 0>= ( n -- f ) 0 >= ;

: within ( n lo hi -- f ) over - >r - r> u< ;
: between ( n lo hi -- f ) 1+ within ;
: bounds ( addr len -- limit start ) over + swap ;

\ --- output helpers ---
: tab tab-char emit ;
: bell bel emit ;
: star [char] * emit ;
: stars ( n -- ) 0 ?do star loop ;
: dots ( n -- ) 0 ?do [char] . emit loop ;
: esc[ esc emit [char] [ emit ;

\ --- ANSI colour (256-colour palette) ---
: fg ( n -- ) esc[ ." 38;5;" 0 .r [char] m emit ;
: bg ( n -- ) esc[ ." 48;5;" 0 .r [char] m emit ;
: normal esc[ ." 0m" ;
: bright esc[ ." 1m" ;

\ --- floats ---
: f2* ( f -- 2f ) fdup f+ ;
: f2/ ( f -- f/2 ) 2.0 f/ ;
: fsq ( f -- f*f ) fdup f* ;
: deg>rad ( f -- f ) pi f* 180.0 f/ ;

\ --- arrays: "10 array counts" then "3 counts [] @" ---
: array ( n "name" -- ) create cells allot ;
: [] ( i addr -- addr' ) swap cells + ;
: barray ( n "name" -- ) create allot ;
: b[] ( i addr -- addr' ) + ;
: 2constant ( n1 n2 "name" -- ) create swap , , does> dup @ swap cell+ @ ;

\ --- string compare ---
: str= ( a1 n1 a2 n2 -- f )
  rot 2dup <> if 2drop 2drop false exit then
  drop
  0 ?do
    over i + c@ over i + c@ <> if 2drop false unloop exit then
  loop
  2drop true ;
