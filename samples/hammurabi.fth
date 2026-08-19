\ hammurabi.fth - the 1968 resource-management classic. Rule Sumeria for ten
\ years without starving your people.
\   forth samples/hammurabi.fth

variable year
variable people
variable grain
variable acres
variable price
variable starved
variable born
variable total-starved
variable plague?
variable harvest
variable eaten-by-rats
variable dead

: ask ( -- n )                       \ read a non-negative number, q quits
  begin
    flush
    pad 32 accept                    ( len )
    dup 0= if cr ." farewell." cr bye then
    pad c@ [char] q = if cr ." farewell." cr bye then
    pad swap number                  ( n f )
    if dup 0 >= if exit then then
    drop
    ." please enter a number (or q to quit): "
  again ;

: .state
  cr ." Hammurabi, I beg to report that in year " year @ .
  cr starved @ . ." people starved and " born @ . ." came to the city." cr
  plague? @ if ." a horrible plague struck! half the people died." cr then
  ." the city population is now " people @ . cr
  ." the city owns " acres @ . ." acres of land." cr
  ." you harvested " harvest @ . ." bushels per acre." cr
  eaten-by-rats @ 0> if ." rats ate " eaten-by-rats @ . ." bushels." cr then
  ." you now have " grain @ . ." bushels in store." cr
  ." land is trading at " price @ . ." bushels per acre." cr ;

: buy-land
  begin
    ." how many acres do you wish to buy? " ask
    dup price @ * grain @ > if
      drop ." O great Hammurabi, we have but " grain @ . ." bushels!" cr
    else
      dup acres +!
      price @ * negate grain +!
      exit
    then
  again ;

: sell-land
  begin
    ." how many acres do you wish to sell? " ask
    dup acres @ > if
      drop ." we own only " acres @ . ." acres." cr
    else
      dup negate acres +!
      price @ * grain +!
      exit
    then
  again ;

: feed-people
  begin
    ." how many bushels shall we feed the people? " ask
    dup grain @ > if
      drop ." we have but " grain @ . ." bushels!" cr
    else
      dup negate grain +!
      exit
    then
  again ;

: plant ( -- acres-planted )
  begin
    ." how many acres shall we plant with seed? " ask
    dup acres @ > if
      drop ." we own only " acres @ . ." acres." cr
    else dup 2/ grain @ > if
      drop ." we do not have enough grain for seed." cr
    else dup people @ 10 * > if
      drop ." each person can tend at most 10 acres." cr
    else
      dup 2/ negate grain +!
      exit
    then then then
  again ;

: rats
  0 eaten-by-rats !
  5 random 2 < if
    grain @ 2 random 3 + / dup eaten-by-rats ! negate grain +!
  then ;

: starvation ( fed -- )              \ bushels actually eaten
  20 / dup people @ >= if
    drop 0 starved !
  else
    people @ swap - dup starved ! dup negate people +!
    dup total-starved +!
    drop
  then ;

: immigration
  starved @ 0= if
    acres @ 20 / grain @ 100 / + people @ / 1+ 5 min
    dup born ! people +!
  else
    0 born !
  then ;

: plague-check
  0 plague? !
  20 random 3 < if
    plague? on
    people @ 2/ negate people +!
  then ;

: one-year
  year incr
  ." --------------------------------------------------" cr
  price @ . ." bushels per acre." cr
  buy-land
  sell-land
  feed-people                        ( fed )
  plant                              ( fed planted )
  \ harvest
  6 random 1+ dup harvest !
  * grain +!
  rats
  starvation
  immigration
  plague-check
  17 random 1+ price ! ;

: verdict
  cr ." --------------------------------------------------" cr
  ." after ten years:" cr
  ." you starved " total-starved @ . ." people in total." cr
  ." you finished with " people @ . ." people and " acres @ . ." acres." cr
  total-starved @ 33 > if
    ." your heavy-handed rule has you deposed and exiled." cr
  else total-starved @ 10 > if
    ." you ruled with an iron fist. the people mutter." cr
  else
    ." a fine reign! the people sing your praises." cr
  then then ;

: intro
  ." ================================================" cr
  ."  HAMMURABI - govern Sumeria for a ten-year term" cr
  ." ================================================" cr
  ." (enter q or a blank line at any prompt to quit)" cr ;

: game
  randomize
  0 year !  100 people !  2800 grain !  1000 acres !  20 price !
  0 total-starved !  0 starved !  5 born !  3 harvest !
  intro
  10 0 do
    one-year
    .state
    people @ 0<= if ." everyone has died. your reign ends in shame." cr leave then
  loop
  people @ 0> if verdict then ;

game
