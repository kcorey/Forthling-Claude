\ sieve.fth - Sieve of Eratosthenes, and the system's speed benchmark.
\   forth samples/sieve.fth

200000 constant limit
create flags limit allot
variable p
variable count
variable shown

: init flags limit 1 fill  0 count ! ;

: mark ( prime -- )          \ strike out its multiples, starting at prime^2
  dup p !
  dup * dup limit < if
    limit swap ?do
      0 flags i + c!
    p @ +loop
  else
    drop
  then ;

: sieve ( -- )
  init
  limit 2 do
    flags i + c@ if i mark then
  loop
  limit 2 do
    flags i + c@ if count incr then
  loop ;

: .primes ( n -- )           \ print the first n primes found
  ." first " dup . ." primes:" cr
  0 shown !
  limit 2 do
    flags i + c@ if
      i 6 .r
      shown incr
      shown @ 10 mod 0= if cr then
      shown @ over >= if leave then
    then
  loop
  drop cr ;

: main
  ." sieving to " limit . cr
  ticks
  sieve
  ticks swap -
  count @ . ." primes found in " . ." ms" cr cr
  20 .primes ;

main
