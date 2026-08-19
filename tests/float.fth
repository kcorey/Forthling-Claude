\ float.fth - the separate floating-point stack

FT{ 1.5 2.5 f+ F-> 4.0 }FT
FT{ 1.5 0.5 f- F-> 1.0 }FT
FT{ 3.0 4.0 f* F-> 12.0 }FT
FT{ 9.0 3.0 f/ F-> 3.0 }FT
FT{ 2.0 fnegate F-> -2.0 }FT
FT{ -2.5 fabs F-> 2.5 }FT
FT{ 16.0 fsqrt F-> 4.0 }FT
FT{ 2.0 3.0 f** F-> 8.0 }FT
FT{ 1.0 2.0 fmin F-> 1.0 }FT
FT{ 1.0 2.0 fmax F-> 2.0 }FT
FT{ 3.7 floor F-> 3.0 }FT
FT{ 3.5 fround F-> 4.0 }FT
FT{ 7 s>f F-> 7.0 }FT
FT{ 1.0 fdup F-> 1.0 1.0 }FT
FT{ 1.0 2.0 fswap F-> 2.0 1.0 }FT
FT{ 1.0 2.0 fover F-> 1.0 2.0 1.0 }FT
FT{ 1.0 2.0 3.0 frot F-> 2.0 3.0 1.0 }FT
FT{ 1.0 2.0 fdrop F-> 1.0 }FT
FT{ 1.0 2.0 fnip F-> 2.0 }FT
FT{ 2.0 f2* F-> 4.0 }FT
FT{ 5.0 f2/ F-> 2.5 }FT
FT{ 3.0 fsq F-> 9.0 }FT
FT{ 1e3 F-> 1000.0 }FT
FT{ -2.5e-1 F-> -0.25 }FT
FT{ 0.0 fsin F-> 0.0 }FT
FT{ 0.0 fcos F-> 1.0 }FT
FT{ 1.0 fln F-> 0.0 }FT
FT{ 0.0 fexp F-> 1.0 }FT
FT{ 100.0 flog F-> 2.0 }FT
FT{ 180.0 deg>rad fsin F-> 0.0 }FT
FT{ 1.0 1.0 fatan2 4.0 f* F-> pi }FT

T{ 1.0 2.0 f< -> true }T
T{ 1.0 2.0 f> -> false }T
T{ 1.0 1.0 f= -> true }T
T{ 1.0 1.0 f<= -> true }T
T{ 1.0 1.0 f>= -> true }T
T{ -1.0 f0< -> true }T
T{ 0.0 f0= -> true }T
T{ 3.9 f>s -> 3 }T
T{ 1.0 2.0 fdepth fdrop fdrop -> 2 }T

variable fvar
FT{ 2.5 fvar f! fvar f@ F-> 2.5 }FT

\ a small floating-point program: Newton's method for a square root
: newton ( f -- root ) 1.0 10 0 do fover fover f/ f+ f2/ loop fnip ;
FT{ 2.0 newton F-> 2.0 fsqrt }FT
