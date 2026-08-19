\ harness.fth - ANS-style test harness:  T{ <code> -> <expected> }T
\ The stack must be empty when T{ runs.

variable #tests
variable #errors
variable actual-depth
variable test-ok
64 array actual-results

: drop-all ( ... -- ) depth 0 ?do drop loop ;

: T{ ( -- ) #tests incr test-ok on
  depth 0<> if ." warning: stack not empty entering test " #tests ? cr drop-all then ;

: -> ( actual... -- ) depth actual-depth !
  actual-depth @ 0 ?do i actual-results [] ! loop ;

: }T ( expected... -- )
  depth actual-depth @ = if
    actual-depth @ 0 ?do i actual-results [] @ <> if test-ok off then loop
  else
    drop-all test-ok off
  then
  test-ok @ 0= if
    #errors incr ." FAILED test #" #tests ? cr
  then ;

\ float variants: FT{ <code> F-> <expected floats> }FT compares to 1e-9
variable factual-depth
64 array factual-results

: FT{ #tests incr test-ok on ;
: F-> fdepth factual-depth !
  factual-depth @ 0 ?do i factual-results [] f! loop ;
: }FT
  fdepth factual-depth @ = if
    factual-depth @ 0 ?do
      i factual-results [] f@ f- fabs 1e-9 f> if test-ok off then
    loop
  else
    fdepth 0 ?do fdrop loop test-ok off
  then
  test-ok @ 0= if
    #errors incr ." FAILED float test #" #tests ? cr
  then ;

: report ( -- )
  ." tests run: " #tests ? ."  errors: " #errors ? cr ;
