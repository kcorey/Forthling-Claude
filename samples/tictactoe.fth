\ tictactoe.fth - you are X, the computer is O and plays perfect minimax.
\   forth samples/tictactoe.fth

create board 9 allot
create lines
  0 c, 1 c, 2 c,   3 c, 4 c, 5 c,   6 c, 7 c, 8 c,
  0 c, 3 c, 6 c,   1 c, 4 c, 7 c,   2 c, 5 c, 8 c,
  0 c, 4 c, 8 c,   2 c, 4 c, 6 c,

variable ln
variable win

: b@ ( i -- v ) board + c@ ;
: b! ( v i -- ) board + c! ;
: empty? ( i -- f ) b@ 0= ;
: other ( p -- p' ) 3 swap - ;

: line-win ( n -- 0|player )
  3 * lines + ln !
  ln @ c@ b@ dup 0= if exit then
  ln @ 1+ c@ b@ over <> if drop 0 exit then
  ln @ 2 + c@ b@ over <> if drop 0 exit then ;

: winner ( -- 0|1|2 )
  0 win !
  8 0 do i line-win ?dup if win ! leave then loop
  win @ ;

: full? ( -- f )
  true 9 0 do i empty? if drop false leave then loop ;

\ negamax-style search: 2 (the computer) maximises, 1 (you) minimises.
: minimax ( player -- score )
  winner ?dup if 2 = if drop 10 else drop -10 then exit then
  full? if drop 0 exit then
  dup 2 = if -1000 else 1000 then    ( player best )
  9 0 do
    i empty? if
      over i b!
      over other recurse             ( player best score )
      2 pick 2 = if max else min then
      0 i b!
    then
  loop
  nip ;

: best-move ( -- i )
  -1 -1000                           ( bestmove best )
  9 0 do
    i empty? if
      2 i b!  1 minimax  0 i b!      ( bestmove best score )
      2dup < if nip nip i swap else drop then
    then
  loop
  drop ;

: .cell ( i -- )
  b@ case
    1 of [char] X endof
    2 of [char] O endof
    [char] . swap
  endcase emit ;

: .board
  cr
  3 0 do
    ."    "
    3 0 do j 3 * i + .cell space loop
    ."     "
    3 0 do j 3 * i + 1+ . loop
    cr
  loop
  cr ;

: get-move ( -- i )
  begin
    ." your move (1-9, q or blank quits): " flush
    pad 16 accept                    ( len )
    dup 0= if cr ." bye." cr bye then
    pad c@ [char] q = if cr ." bye." cr bye then
    pad swap number                  ( n f )
    if
      1- dup 0 9 within if dup empty? if exit then then
    then
    drop
    ." that square is not free." cr
  again ;

: game-over? ( -- f )
  winner ?dup if
    .board
    1 = if ." you win!" else ." I win." then cr
    true exit
  then
  full? if .board ." a draw." cr true exit then
  false ;

: play
  board 9 erase
  ." tic-tac-toe - you are X, I am O. Squares are numbered 1-9." cr
  begin
    .board
    get-move 1 swap b!
    game-over? if exit then
    ." thinking..." cr
    best-move 2 swap b!
    game-over? if exit then
  again ;

play
