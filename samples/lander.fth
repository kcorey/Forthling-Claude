\ lander.fth - a real-time lunar lander in colour, sized to your terminal.
\   forth samples/lander.fth          play it
\   forth samples/lander.fth auto     watch the autopilot fly it
\
\   a / d  or  left / right      side thrusters
\   w / space / up               main engine
\   q                            quit
\
\ Lunar gravity is gentle but relentless. Touch down on one of the green pads,
\ slowly and squarely. The score rewards landing near the centre of a pad,
\ touching down softly, and finishing with fuel to spare: narrow pads are
\ worth more. The scene fills the window, and a taller window means a longer
\ fall, so the tanks are filled to match.
\
\ The pale arrows around the module show momentum: a stack of v below you is
\ how fast you are falling, ^ above is how fast you are climbing, < and > are
\ sideways drift. They turn pink once they exceed what the legs can take, so
\ you can watch a burn take effect.

400 constant MAXW
200 constant MAXH
MAXW MAXH * constant MAXCELLS
240 constant MAXSTARS
18 constant DEBRIS

\ ---------------------------------------------------------------- palette
244 constant c-rock
 46 constant c-pad
226 constant c-ship
202 constant c-flame
 39 constant c-hud
196 constant c-bad
 33 constant c-sky
240 constant c-star
 82 constant c-good
255 constant c-gauge               \ pale arrows: safe
217 constant c-gauge-hot           \ pale pink arrows: too fast
195 constant c-gauge-lift          \ pale blue arrows: climbing
252 constant c-panel

\ ---------------------------------------------------------------- storage
create scr    MAXCELLS allot       \ character buffer
create clr    MAXCELLS allot       \ colour buffer
create ground MAXW allot           \ terrain top row, per column
create padid  MAXW allot           \ 0 = rock, otherwise the pad number

4 array padcx                      \ pad centre column
4 array padhw                      \ pad half width
4 array padmul                     \ pad score multiplier

MAXSTARS array starx
MAXSTARS array stary
variable nstars

DEBRIS array dbx                   \ exploding module parts (floats)
DEBRIS array dby
DEBRIS array dbvx
DEBRIS array dbvy
DEBRIS array dbch
DEBRIS array dbrest

variable scr-w                     \ playfield width, the whole window
variable scr-h                     \ playfield height, the window minus the HUD
: W scr-w @ ;
: H scr-h @ ;

fvariable lx                       \ position, in character cells
fvariable ly
fvariable vx                       \ velocity, cells per second
fvariable vy
fvariable dt
fvariable grav
fvariable up-thrust
fvariable side-thrust

variable fuel
variable fuel0
variable flying
variable outcome                   \ 1 = landed, 2 = crashed, 3 = quit
variable burning
variable side-burn
variable score
variable bonus-pad
variable frames
variable curcol
variable gh
variable cx0
variable gy0
variable sx
variable sy
variable dist
variable halfw
variable pn
variable pwidth
variable pleft
variable pheight
variable pregion
variable fl-h                      \ current flame height, 0 when coasting

3 constant UP-COST
2 constant SIDE-COST
2.5 fconstant SAFE-VY              \ fastest survivable descent
1.5 fconstant SAFE-VX              \ fastest survivable drift

\ ---------------------------------------------------------------- window
: measure-window
  term-size                        ( cols rows )
  3 - MAXH min 12 max scr-h !
  MAXW min 40 max scr-w ! ;

\ ---------------------------------------------------------------- screen
: at ( x y -- offset ) W * + ;
: put ( ch colour x y -- )
  dup 0 H 1- between 0= if 2drop 2drop exit then
  over 0 W 1- between 0= if 2drop 2drop exit then
  at dup >r clr + c! r> scr + c! ;

: clear-screen
  scr W H * bl fill
  clr W H * c-sky fill ;

: setcol ( n -- ) dup curcol @ <> if dup fg then curcol ! ;

: show-screen
  0 0 at-xy
  H 0 do
    W 0 do
      i j at dup clr + c@ setcol scr + c@ emit
    loop
    normal -1 curcol ! cr
  loop ;

\ ---------------------------------------------------- text into the screen
variable psx
variable psy
variable pcol

: put-str ( addr len colour x y -- )
  psy ! psx ! pcol !
  0 ?do
    dup i + c@ pcol @ psx @ i + psy @ put
  loop
  drop ;

: put-centre ( addr len colour y -- )
  psy ! pcol !
  dup W swap - 2/                  ( addr len x )
  pcol @ swap psy @ put-str ;

\ string building, so panel lines can mix words and numbers
create linebuf 200 allot
create numbuf 24 allot
variable lp
variable nv
variable np

: n>buf ( n -- addr len )
  nv ! 0 np !
  nv @ 0= if
    [char] 0 numbuf c! 1 np !
  else
    nv @ 0< nv @ abs nv !          ( negative? )
    begin nv @ 0> while
      nv @ 10 mod [char] 0 + numbuf np @ + c!
      np incr
      nv @ 10 / nv !
    repeat
    if [char] - numbuf np @ + c! np incr then
    np @ 2/ 0 do
      numbuf i + c@
      numbuf np @ 1- i - + c@
      numbuf i + c!
      numbuf np @ 1- i - + c!
    loop
  then
  numbuf np @ ;

: line0 0 lp ! ;
: +s ( addr len -- ) linebuf lp @ + swap dup >r move r> lp +! ;
: +n ( n -- ) n>buf +s ;
: line ( -- addr len ) linebuf lp @ ;

\ ---------------------------------------------------------------- terrain
: lowest  ( -- row ) H 2 - ;             \ deepest terrain row
: highest ( -- row ) H H 5 / - 4 - ;     \ tallest peak, scaled to the window

: new-terrain
  H 4 - gh !
  W 0 do
    3 random 1- gh +!
    gh @ highest max lowest min gh !
    gh @ ground i + c!
    0 padid i + c!
  loop ;

: carve-pad                        \ uses pn, pwidth, pleft
  ground pleft @ + c@ pheight !
  pwidth @ 0 do
    pheight @ ground pleft @ i + + c!
    pn @      padid  pleft @ i + + c!
  loop
  pleft @ pwidth @ 2/ +  pn @ padcx []  !
  pwidth @ 2/            pn @ padhw []  !
  pn @                   pn @ padmul [] ! ;

: pad-slot                         \ pad pn, width pwidth, inside region pregion
  W 3 / pregion @ * 2 +
  W 3 / pwidth @ - 4 - 1 max random +
  pleft ! carve-pad ;

: pads
  1 pn ! W  8 / 6 max pwidth ! 0 pregion ! pad-slot
  2 pn ! W 11 / 5 max pwidth ! 1 pregion ! pad-slot
  3 pn ! W 18 / 3 max pwidth ! 2 pregion ! pad-slot ;

: new-stars
  W H * 45 / MAXSTARS min 8 max nstars !
  nstars @ 0 do
    W random                 i starx [] !
    highest 2 - 1 max random i stary [] !
  loop ;

: draw-stars
  nstars @ 0 do
    [char] . c-star  i starx [] @  i stary [] @  put
  loop ;

: draw-col ( x -- )
  cx0 !
  cx0 @ ground + c@ gy0 !
  cx0 @ padid + c@ if
    [char] = c-pad cx0 @ gy0 @ put
  else
    [char] ^ c-rock cx0 @ gy0 @ put
  then
  H gy0 @ 1+ ?do
    [char] # c-rock cx0 @ i put
  loop ;

: draw-terrain W 0 do i draw-col loop ;

: draw-pad-labels                  \ the multiplier, printed above each pad
  4 1 do
    i padmul [] @ [char] 0 +
    c-pad
    i padcx [] @
    dup ground + c@ 1-
    put
  loop ;

\ ---------------------------------------------------------------- the ship
: ship-cells lx f@ f>s sx !  ly f@ f>s sy ! ;

: draw-hull
  [char] / c-ship sx @ 1- sy @ 1- put
  [char] ^ c-ship sx @    sy @ 1- put
  [char] \ c-ship sx @ 1+ sy @ 1- put
  [char] | c-ship sx @ 1- sy @    put
  [char] o c-ship sx @    sy @    put
  [char] | c-ship sx @ 1+ sy @    put ;

\ ------------------------------------------------------------- the exhaust
: flame-char ( -- c )
  4 random case
    0 of [char] * endof
    1 of [char] ^ endof
    2 of [char] # endof
    [char] , swap
  endcase ;

: flame-colour ( n -- c )          \ n counts down the plume from the nozzle
  case
    0 of 227 endof
    1 of 220 endof
    2 of 214 endof
    3 of 208 endof
    196 swap
  endcase ;

: draw-flame
  0 fl-h !
  burning @ 0= if exit then
  4 3 random + fl-h !              \ 4 to 6 cells, flickering every frame
  fl-h @ 0 do
    flame-char  i flame-colour  sx @  sy @ 1+ i +  put
  loop
  [char] ( 220 sx @ 1- sy @ 1+ put
  [char] ) 220 sx @ 1+ sy @ 1+ put
  2 random 0= if [char] * 208 sx @ 1- sy @ 2 + put then
  2 random 0= if [char] * 208 sx @ 1+ sy @ 2 + put then
  fl-h @ 4 > if
    [char] . 196 sx @ 1- sy @ fl-h @ + put
    [char] . 196 sx @ 1+ sy @ fl-h @ + put
  then ;

\ -------------------------------------------------------- momentum gauges
\ One pale arrow per half a cell per second, pointing the way you are moving.
: vy-colour ( -- c ) vy f@ fabs SAFE-VY f< if c-gauge else c-gauge-hot then ;
: vx-colour ( -- c ) vx f@ fabs SAFE-VX f< if c-gauge else c-gauge-hot then ;
: arrows ( -- n ) ( F: speed -- ) fabs 2.0 f* f>s 9 min ;

: draw-fall  ( n -- )              \ below the ship, clear of the exhaust
  0 ?do [char] v vy-colour sx @ sy @ 2 + fl-h @ + i + put loop ;
: draw-climb ( n -- )
  0 ?do [char] ^ c-gauge-lift sx @ sy @ 2 - i - put loop ;
: draw-right ( n -- )
  0 ?do [char] > vx-colour sx @ 3 + i + sy @ put loop ;
: draw-left  ( n -- )
  0 ?do [char] < vx-colour sx @ 3 - i - sy @ put loop ;

\ FDUP keeps the speed on the float stack so ARROWS can count it and F0< can
\ then ask which way it points.
: draw-gauges
  vy f@ fdup arrows                          ( n ; F: vy )
  dup 0= if drop fdrop else
    f0< if draw-climb else draw-fall then
  then
  vx f@ fdup arrows                          ( n ; F: vx )
  dup 0= if drop fdrop else
    f0< if draw-left else draw-right then
  then ;

: draw-ship draw-hull draw-flame draw-gauges ;

\ ---------------------------------------------------------------- physics
: use-fuel ( n -- ) fuel @ swap - 0max fuel ! ;

: apply-fuel
  fuel @ 0<= if 0 burning ! 0 side-burn ! exit then
  burning @ if UP-COST use-fuel then
  side-burn @ if SIDE-COST use-fuel then ;

: physics
  vy f@ grav f@ dt f@ f* f+ vy f!
  burning @ if
    vy f@ up-thrust f@ dt f@ f* f- vy f!
  then
  side-burn @ dup 0<> if
    s>f side-thrust f@ f* dt f@ f* vx f@ f+ vx f!
  else drop then
  lx f@ vx f@ dt f@ f* f+ lx f!
  ly f@ vy f@ dt f@ f* f+ ly f! ;

: bounds
  lx f@ 2.0 f< if 2.0 lx f! 0.0 vx f! then
  lx f@ W 3 - s>f f> if W 3 - s>f lx f! 0.0 vx f! then
  ly f@ 1.0 f< if 1.0 ly f! 0.0 vy f! then ;

\ ---------------------------------------------------------------- landing
: ground-at ( x -- y ) 0max W 1- min ground + c@ ;
: pad-at    ( x -- n ) 0max W 1- min padid + c@ ;

: contact? ( -- f )
  sy @ 1+
  sx @ 1- ground-at
  sx @    ground-at min
  sx @ 1+ ground-at min
  >= ;

: level? ( -- f )
  sx @ ground-at
  dup sx @ 1- ground-at =
  swap sx @ 1+ ground-at =
  and ;

: on-pad ( -- n )
  sx @ pad-at
  dup 0= if exit then
  dup sx @ 1- pad-at <> if drop 0 exit then
  dup sx @ 1+ pad-at <> if drop 0 exit then ;

: gentle? ( -- f )
  vy f@ SAFE-VY f<
  vx f@ fabs SAFE-VX f< and ;

: centre-score ( pad -- n )
  dup padcx [] @ lx f@ f>s - abs dist !
  padhw [] @ 1 max halfw !
  halfw @ dist @ - 0max 100 * halfw @ / ;

: soft-score ( -- n )
  vy f@ fabs 40.0 f* f>s 100 swap - 0max ;

: tally ( pad -- )
  dup bonus-pad !
  dup centre-score
  soft-score +
  swap padmul [] @ *
  fuel @ 10 / +
  score ! ;

: check-landing
  contact? 0= if exit then
  0 flying !
  on-pad ?dup 0= if 2 outcome ! 0 score ! exit then
  level?  0= if drop 2 outcome ! 0 score ! exit then
  gentle? 0= if drop 2 outcome ! 0 score ! exit then
  1 outcome !
  tally ;

\ ---------------------------------------------------------------- input
: drain-keys begin key? while key drop repeat ;

: thrust-up    1 burning ! ;
: thrust-left -1 side-burn ! ;
: thrust-right 1 side-burn ! ;

: arrow-key
  key? 0= if exit then
  key drop
  key? 0= if exit then
  key case
    [char] A of thrust-up endof
    [char] C of thrust-right endof
    [char] D of thrust-left endof
  endcase ;

: read-keys
  0 burning ! 0 side-burn !
  begin key? while
    key case
      [char] w of thrust-up endof
      bl       of thrust-up endof
      [char] a of thrust-left endof
      [char] d of thrust-right endof
      [char] q of 0 flying ! 3 outcome ! endof
      27       of arrow-key endof
    endcase
  repeat ;

\ ---------------------------------------------------------------- display
: .arrows ( n ch -- ) swap 0 ?do dup emit loop drop ;

: .descent
  vy-colour fg
  vy f@ arrows dup 0= if drop ." -" else
    vy f@ f0< if [char] ^ else [char] v then .arrows
  then
  normal ;

: .drift
  vx-colour fg
  vx f@ arrows dup 0= if drop ." -" else
    vx f@ f0< if [char] < else [char] > then .arrows
  then
  normal ;

: fuel-colour
  fuel @ fuel0 @ 3 / > if c-good else
  fuel @ fuel0 @ 8 / > if c-hud else c-bad then then fg ;

: hud
  c-hud fg
  ." alt " sx @ ground-at s>f ly f@ f- 5 1 f.r
  ."   fuel " fuel-colour fuel @ 5 .r
  c-hud fg ."   time " frames @ 20 / 3 .r ." s"
  normal cr
  c-hud fg ." descent " vy f@ fabs 5 2 f.r space .descent
  c-hud fg ."    drift " vx f@ fabs 5 2 f.r space .drift
  normal cr
  c-hud fg
  W 64 > if
    ." a/d or arrows steer   w or space burns   q quits   pink arrows = too fast"
  else
    ." a/d steer  w burns  q quits"
  then
  normal flush ;

: draw-world
  clear-screen draw-stars draw-terrain draw-pad-labels ;

: render
  draw-world ship-cells draw-ship
  show-screen hud ;

\ ---------------------------------------------------------------- break-up
\ On a crash the module comes apart and its letters are thrown across the
\ surface, where they stay.
: debris-char ( i -- c ) s" /^\|o|*+.,'" drop swap 11 mod + c@ ;

: scatter                          \ seed the pieces from the ship's position
  DEBRIS 0 do
    lx f@ i 3 mod s>f f+ 1.0 f- i dbx [] f!
    ly f@ i 3 / s>f 2.0 f/ f-      i dby [] f!
    9 random 4 - s>f 1.6 f*        i dbvx [] f!
    5 random 1+ s>f -2.2 f*        i dbvy [] f!
    i debris-char                  i dbch [] !
    0                              i dbrest [] !
  loop ;

: fly-debris                       \ one step of the break-up animation
  DEBRIS 0 do
    i dbrest [] @ 0= if
      i dbvy [] f@ 7.0 0.08 f* f+ i dbvy [] f!
      i dbx [] f@ i dbvx [] f@ 0.08 f* f+ i dbx [] f!
      i dby [] f@ i dbvy [] f@ 0.08 f* f+ i dby [] f!
      i dbx [] f@ f>s 0 W 1- between 0= if
        0.0 i dbvx [] f!
      then
      i dby [] f@ f>s  i dbx [] f@ f>s ground-at 1- >= if
        i dbx [] f@ f>s ground-at 1- s>f i dby [] f!
        1 i dbrest [] !
      then
    then
  loop ;

: draw-debris
  DEBRIS 0 do
    i dbch [] @
    i dbrest [] @ if c-rock else 202 i 3 mod 6 * + then
    i dbx [] f@ f>s
    i dby [] f@ f>s
    put
  loop ;

: explode
  scatter
  40 0 do
    draw-world draw-debris show-screen
    c-bad fg ." BREAK-UP" normal cr cr flush
    fly-debris
    40 ms
  loop ;

\ ------------------------------------------------------------- end-of-game
\ The panel is drawn into the middle of the last frame, and the keyboard is
\ drained first so a thruster key still in flight cannot dismiss it.
variable bx
variable by
variable bw
variable bh

: draw-panel-box ( w h -- )
  bh ! bw !
  W bw @ - 2/ bx !  H bh @ - 2/ by !
  bh @ 0 do
    bw @ 0 do
      bl c-panel bx @ i + by @ j + put
    loop
  loop
  bw @ 0 do
    [char] - c-panel bx @ i + by @ put
    [char] - c-panel bx @ i + by @ bh @ 1- + put
  loop
  bh @ 0 do
    [char] | c-panel bx @        by @ i + put
    [char] | c-panel bx @ bw @ 1- + by @ i + put
  loop ;

: landed-panel
  line0 s" TOUCHDOWN - the Eagle has landed" +s
  line c-good by @ 1+ put-centre
  line0 s" pad " +s bonus-pad @ +n s"  (x" +s bonus-pad @ padmul [] @ +n
  s" )   centre " +s bonus-pad @ centre-score +n s" /100   softness " +s
  soft-score +n s" /100" +s
  line c-panel by @ 3 + put-centre
  line0 s" fuel left " +s fuel @ +n s"  of " +s fuel0 @ +n
  line c-panel by @ 4 + put-centre
  line0 s" FINAL SCORE " +s score @ +n
  line c-good by @ 5 + put-centre ;

: crash-panel
  line0 s" CRASH - the module is scattered across the regolith" +s
  line c-bad by @ 1+ put-centre
  line0
  vy f@ SAFE-VY f>= if s" you came down too fast" +s
  else vx f@ fabs SAFE-VX f>= if s" too much sideways drift" +s
  else on-pad 0= if s" that was not a landing pad" +s
  else s" the ground there was not level" +s
  then then then
  line c-panel by @ 3 + put-centre
  line0 s" SCORE 0" +s
  line c-bad by @ 5 + put-centre ;

: quit-panel
  line0 s" MISSION ABORTED" +s
  line c-panel by @ 1+ put-centre ;

: panel
  W 6 - 60 min 30 max 9 draw-panel-box
  outcome @ case
    1 of landed-panel endof
    2 of crash-panel endof
    3 of quit-panel endof
  endcase
  line0 s" R = fly again      Q or ENTER = end" +s
  line c-hud by @ 7 + put-centre
  show-screen
  cr flush ;

variable decided
variable again-flag

: wait-decision ( -- f )           \ true = fly again
  300 ms drain-keys                \ swallow keys pressed during the crash
  0 decided ! 0 again-flag !
  begin decided @ 0= while
    key
    dup 0< if
      drop 1 decided !             \ end of input: stop
    else
      dup 27 = if
        drop 30 ms drain-keys      \ an arrow key still in flight: ignore it
      else
        dup [char] r = over [char] R = or if
          1 again-flag ! 1 decided !
        else
          dup [char] q = over [char] Q = or
          over 13 = or over 10 = or if 1 decided ! then
        then
        drop
      then
    then
  repeat
  again-flag @ 0<> ;

\ -------------------------------------------------------------- autopilot
fvariable want-vx
fvariable want-vy
fvariable dx0
variable tgt

: alt ( -- ) ( F: -- altitude ) sx @ ground-at s>f ly f@ f- ;

: nearest-pad ( -- n )
  1 tgt !
  4 1 do
    i     padcx [] @ lx f@ f>s - abs
    tgt @ padcx [] @ lx f@ f>s - abs
    < if i tgt ! then
  loop
  tgt @ ;

: autopilot
  0 burning ! 0 side-burn !
  nearest-pad padcx [] @ s>f lx f@ f- dx0 f!
  dx0 f@ 0.7 f* -3.0 fmax 3.0 fmin want-vx f!
  vx f@ want-vx f@ 0.15 f- f< if  1 side-burn ! then
  vx f@ want-vx f@ 0.15 f+ f> if -1 side-burn ! then
  dx0 f@ fabs 1.2 f> if
    ly f@ 4.0 f> if -1.2 else 0.4 then want-vy f!
  else
    alt 0.30 f* 0.5 f+ 2.0 fmin want-vy f!
  then
  vy f@ want-vy f@ f> if 1 burning ! then ;

\ ---------------------------------------------------------------- game
: refuel                           \ a taller window means a longer fall
  300 H 34 * + dup fuel0 ! fuel ! ;

: new-game
  measure-window
  new-terrain pads new-stars
  W 2/ 8 - 16 random + s>f lx f!
  2.0 ly f!
  5 random 2 - s>f 6.0 f/ vx f!
  0.0 vy f!
  0.05 dt f!
  1.2 grav f!                      \ gentle lunar gravity: time to react
  3.2 up-thrust f!
  2.2 side-thrust f!
  refuel
  0 score ! 0 bonus-pad ! 0 frames !
  0 fl-h ! 0 burning ! 0 side-burn !
  0 outcome ! flying on ;

: show-wreck  draw-world draw-debris ;
: show-intact draw-world ship-cells draw-ship ;

: finish                           \ break-up, then the panel, then a decision
  outcome @ 2 = if
    explode show-wreck             \ the pieces stay where they fell
  else
    show-intact
  then
  panel
  wait-decision ;

: fly-manual
  raw-on cursor-off
  begin
    page drain-keys new-game
    begin flying @ while
      read-keys apply-fuel physics bounds ship-cells check-landing
      render
      frames incr
      50 ms
    repeat
    outcome @ 3 = if false else finish then
  while repeat
  cursor-on raw-off page
  outcome @ 3 = if ." mission aborted." cr then ;

: fly-auto                         \ "lander.fth auto" flies itself
  cursor-off page new-game
  begin flying @ while
    autopilot apply-fuel physics bounds ship-cells check-landing
    render
    frames incr
    30 ms
  repeat
  outcome @ 2 = if explode show-wreck else show-intact then
  panel cursor-on cr ;

: main
  randomize
  argc 0> if
    0 arg s" auto" str= if fly-auto exit then
  then
  fly-manual ;

main
