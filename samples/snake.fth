\ snake.fth - the arcade classic, driven by non-blocking KEY?.
\   forth samples/snake.fth
\   w a s d or the arrow keys steer, q quits.

40 constant W
20 constant H
W H * constant CELLS

create bodyx CELLS allot            \ ring buffer of body segments
create bodyy CELLS allot
create board CELLS allot

variable head
variable len
variable dx
variable dy
variable fx
variable fy
variable score
variable playing
variable hx
variable hy
variable hitf
variable nx
variable ny

: ring ( i -- i ) CELLS + CELLS mod ;
: seg-x ( i -- x ) ring bodyx + c@ ;
: seg-y ( i -- y ) ring bodyy + c@ ;
: put-seg ( x y i -- ) ring dup >r bodyy + c! r> bodyx + c! ;

: hit? ( x y -- f )                 \ does the snake occupy this square?
  hy ! hx ! hitf off
  len @ 0 do
    head @ i - dup seg-x hx @ = swap seg-y hy @ = and if hitf on leave then
  loop
  hitf @ ;

: new-food
  begin
    W random fx !  H random fy !
    fx @ fy @ hit? 0=
  until ;

: render
  board CELLS bl fill
  len @ 0 do
    head @ i - dup seg-x swap seg-y W * +
    board + [char] o swap c!
  loop
  head @ seg-x head @ seg-y W * + board + [char] @ swap c!
  fx @ fy @ W * + board + [char] * swap c!
  0 0 at-xy
  ." +" W dots ." +" cr
  H 0 do
    [char] | emit
    W 0 do j W * i + board + c@ emit loop
    [char] | emit cr
  loop
  ." +" W dots ." +" cr
  ." score " score @ . ."   length " len @ . ."   (q quits)   " flush ;

: turn ( c -- )
  case
    [char] w of dy @ 1 = if exit then 0 dx ! -1 dy ! endof
    [char] s of dy @ -1 = if exit then 0 dx ! 1 dy ! endof
    [char] a of dx @ 1 = if exit then -1 dx ! 0 dy ! endof
    [char] d of dx @ -1 = if exit then 1 dx ! 0 dy ! endof
    [char] q of playing off endof
    27 of
      key drop key
      case
        [char] A of dy @ 1 = if exit then 0 dx ! -1 dy ! endof
        [char] B of dy @ -1 = if exit then 0 dx ! 1 dy ! endof
        [char] C of dx @ -1 = if exit then 1 dx ! 0 dy ! endof
        [char] D of dx @ 1 = if exit then -1 dx ! 0 dy ! endof
      endcase
    endof
  endcase ;

: input begin key? while key turn repeat ;

: step
  head @ seg-x dx @ + nx !
  head @ seg-y dy @ + ny !
  nx @ 0 W within 0= if playing off exit then
  ny @ 0 H within 0= if playing off exit then
  nx @ ny @ hit? if playing off exit then
  nx @ fx @ = ny @ fy @ = and if
    10 score +!
    len incr
    new-food
  then
  head @ 1+ head !
  nx @ ny @ head @ put-seg ;

: init
  randomize
  board CELLS bl fill
  0 head !  3 len !  1 dx !  0 dy !  0 score !  playing on
  10 10 0 put-seg
  9 10 -1 put-seg
  8 10 -2 put-seg
  new-food ;

: play
  init raw-on cursor-off page
  begin playing @ while
    render
    input
    step
    90 ms
  repeat
  render cr
  ." game over - final score " score @ . cr
  cursor-on raw-off ;

play
