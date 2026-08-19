\ mandelbrot.fth - a Mandelbrot explorer in colour, with pan, zoom and rotate.
\   forth samples/mandelbrot.fth
\
\   w a s d / arrow keys   pan          + -   zoom in and out
\   [ ]                    rotate       i o   more / fewer iterations
\   r                      reset        q     quit

78 constant PW
30 constant PH

variable maxit
fvariable ctrx
fvariable ctry
fvariable span
fvariable angle
fvariable ca
fvariable sa
fvariable ustep
fvariable u
fvariable v
fvariable cre
fvariable cim
fvariable zx
fvariable zy
fvariable tmp
fvariable mx
fvariable my
variable iter

: reset
  -0.5 ctrx f!  0.0 ctry f!  3.0 span f!  0.0 angle f!  120 maxit ! ;

: recompute
  span f@ PW s>f f/ ustep f!
  angle f@ fcos ca f!
  angle f@ fsin sa f! ;

: pixel ( i j -- )                  \ set cre/cim for screen cell i,j
  s>f PH 2/ s>f f- ustep f@ f* f2* v f!
  s>f PW 2/ s>f f- ustep f@ f* u f!
  u f@ ca f@ f* v f@ sa f@ f* f- ctrx f@ f+ cre f!
  u f@ sa f@ f* v f@ ca f@ f* f+ ctry f@ f+ cim f! ;

: escape ( -- n )                   \ iterations before |z| leaves radius 2
  0.0 zx f!  0.0 zy f!  0 iter !
  begin
    zx f@ fsq zy f@ fsq f+ 4.0 f<
    iter @ maxit @ < and
  while
    zx f@ fsq zy f@ fsq f- cre f@ f+ tmp f!
    zx f@ zy f@ f* f2* cim f@ f+ zy f!
    tmp f@ zx f!
    iter incr
  repeat
  iter @ ;

: ink ( n -- )                      \ colour by escape time
  dup maxit @ >= if
    drop 16 bg
  else
    3 * 210 mod 17 + bg
  then ;

: draw
  0 0 at-xy
  PH 0 do
    PW 0 do
      i j pixel escape ink space
    loop
    normal cr
  loop ;

: status
  normal
  ." centre " ctrx f@ 12 6 f.r ctry f@ 12 6 f.r ."   span " span f@ 12 8 f.r
  ."   angle " angle f@ 6 2 f.r ."   iters " maxit @ . cr
  ." wasd pan  +/- zoom  [ ] rotate  i/o iters  r reset  q quit    " flush ;

: pan                               \ move by (mx,my) in rotated screen axes
  mx f@ ca f@ f* my f@ sa f@ f* f- ctrx f@ f+ ctrx f!
  mx f@ sa f@ f* my f@ ca f@ f* f+ ctry f@ f+ ctry f! ;

: step-size span f@ 0.15 f* ;
: pan-left  step-size fnegate mx f!  0.0 my f! pan ;
: pan-right step-size mx f!          0.0 my f! pan ;
: pan-up    0.0 mx f! step-size fnegate my f! pan ;
: pan-down  0.0 mx f! step-size my f! pan ;
: zoom-in   span f@ 0.7 f* span f!  maxit @ 20 + 400 min maxit ! ;
: zoom-out  span f@ 1.4 f* span f!  maxit @ 20 - 40 max maxit ! ;
: turn-cw   angle f@ 0.15 f+ angle f! ;
: turn-ccw  angle f@ 0.15 f- angle f! ;
: more-iter maxit @ 40 + 2000 min maxit ! ;
: less-iter maxit @ 40 - 40 max maxit ! ;

: leave-app raw-off cursor-on normal page bye ;

: arrow ( -- )                      \ an escape sequence: ESC [ A/B/C/D
  key drop key
  case
    [char] A of pan-up endof
    [char] B of pan-down endof
    [char] C of pan-right endof
    [char] D of pan-left endof
  endcase ;

: handle ( c -- )
  case
    [char] w of pan-up endof
    [char] s of pan-down endof
    [char] a of pan-left endof
    [char] d of pan-right endof
    [char] + of zoom-in endof
    [char] = of zoom-in endof
    [char] - of zoom-out endof
    [char] ] of turn-cw endof
    [char] [ of turn-ccw endof
    [char] i of more-iter endof
    [char] o of less-iter endof
    [char] r of reset endof
    [char] q of leave-app endof
    27 of arrow endof
  endcase ;

: explore
  reset raw-on cursor-off page
  begin
    recompute draw status
    key dup 0< if drop leave-app then
    handle
  again ;

explore
