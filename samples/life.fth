\ life.fth - Conway's Game of Life on a wrapping torus, with a glider gun.
\   forth samples/life.fth [generations]
\ Press any key to stop early.

64 constant W
22 constant H
W H * constant SIZE

create grid SIZE allot
create nxt  SIZE allot

variable cx
variable cy
variable cnt

: idx ( x y -- offset ) W * + ;
: wrapx ( x -- x ) W + W mod ;
: wrapy ( y -- y ) H + H mod ;
: g@ ( x y -- 0|1 ) wrapy swap wrapx swap idx grid + c@ ;
: n! ( v x y -- ) idx nxt + c! ;
: set ( x y -- ) idx grid + 1 swap c! ;

: neighbours ( x y -- n )
  cy ! cx ! 0 cnt !
  2 -1 do
    2 -1 do
      i j or if
        cx @ i + cy @ j + g@ cnt +!
      then
    loop
  loop
  cnt @ ;

: rule ( alive n -- v )
  swap if dup 2 = swap 3 = or else 3 = then 1 and ;

: step
  H 0 do
    W 0 do
      i j g@  i j neighbours rule
      i j n!
    loop
  loop
  nxt grid SIZE move ;

: show
  0 0 at-xy
  H 0 do
    W 0 do
      i j g@ if [char] O else bl then emit
    loop
    cr
  loop ;

\ Gosper glider gun, offset a little from the corner
: gun
  1 5 set  1 6 set  2 5 set  2 6 set
  11 5 set 11 6 set 11 7 set
  12 4 set 12 8 set
  13 3 set 13 9 set  14 3 set 14 9 set
  15 6 set
  16 4 set 16 8 set
  17 5 set 17 6 set 17 7 set
  18 6 set
  21 3 set 21 4 set 21 5 set
  22 3 set 22 4 set 22 5 set
  23 2 set 23 6 set
  25 1 set 25 2 set 25 6 set 25 7 set
  35 3 set 35 4 set 36 3 set 36 4 set ;

: sow
  grid SIZE erase
  gun
  \ a sprinkling of random cells in the lower half
  200 0 do
    W random  H 2/ H 2/ random +  set
  loop ;

: gens ( -- n )
  argc 0> if 0 arg number if exit then drop then 120 ;

: run ( n -- )
  raw-on cursor-off page
  0 do
    show
    ." generation " i 1+ . ."  (any key stops)   " cr
    step
    50 ms
    key? if key drop leave then
  loop
  cursor-on raw-off cr ;

12345 seed
sow
gens run
