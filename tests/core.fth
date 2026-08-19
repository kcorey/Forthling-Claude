\ core.fth - stack, arithmetic, comparison and memory words

\ --- stack ---
T{ 1 2 3 -> 1 2 3 }T
T{ 1 dup -> 1 1 }T
T{ 1 2 drop -> 1 }T
T{ 1 2 swap -> 2 1 }T
T{ 1 2 over -> 1 2 1 }T
T{ 1 2 nip -> 2 }T
T{ 1 2 tuck -> 2 1 2 }T
T{ 1 2 3 rot -> 2 3 1 }T
T{ 1 2 3 -rot -> 3 1 2 }T
T{ 0 ?dup -> 0 }T
T{ 9 ?dup -> 9 9 }T
T{ 1 2 2dup -> 1 2 1 2 }T
T{ 1 2 3 4 2drop -> 1 2 }T
T{ 1 2 3 4 2swap -> 3 4 1 2 }T
T{ 1 2 3 4 2over -> 1 2 3 4 1 2 }T
T{ 5 6 7 1 pick -> 5 6 7 6 }T
T{ 1 2 3 2 roll -> 2 3 1 }T
T{ 1 2 3 depth -> 1 2 3 3 }T
T{ 1 2 3 3dup -> 1 2 3 1 2 3 }T

\ --- return stack ---
: rs1 5 >r 1 r> ;
T{ rs1 -> 1 5 }T
: rs2 7 >r r@ r> ;
T{ rs2 -> 7 7 }T
: rs3 1 2 2>r 9 2r> ;
T{ rs3 -> 9 1 2 }T

\ --- arithmetic ---
T{ 2 3 + -> 5 }T
T{ 10 3 - -> 7 }T
T{ 6 7 * -> 42 }T
T{ 20 5 / -> 4 }T
T{ -7 2 / -> -4 }T
T{ -7 2 mod -> 1 }T
T{ 7 -2 / -> -4 }T
T{ 17 5 /mod -> 2 3 }T
T{ 10 3 2 */ -> 15 }T
T{ 5 negate -> -5 }T
T{ -9 abs -> 9 }T
T{ 3 7 min -> 3 }T
T{ 3 7 max -> 7 }T
T{ 5 1+ -> 6 }T
T{ 5 1- -> 4 }T
T{ 5 2* -> 10 }T
T{ 5 2/ -> 2 }T
T{ 1 4 lshift -> 16 }T
T{ 16 4 rshift -> 1 }T
T{ 12 10 and -> 8 }T
T{ 12 10 or -> 14 }T
T{ 12 10 xor -> 6 }T
T{ 0 invert -> -1 }T
T{ 4 sq -> 16 }T
T{ -3 0max -> 0 }T

\ --- comparison ---
T{ 3 3 = -> true }T
T{ 3 4 = -> false }T
T{ 3 4 <> -> true }T
T{ 3 4 < -> true }T
T{ 4 3 > -> true }T
T{ 3 3 <= -> true }T
T{ 3 3 >= -> true }T
T{ 1 -1 u< -> true }T
T{ -1 1 u> -> true }T
T{ 0 0= -> true }T
T{ 1 0<> -> true }T
T{ -1 0< -> true }T
T{ 1 0> -> true }T
T{ 0 0<= -> true }T
T{ 0 0>= -> true }T
T{ 1 0<= -> false }T
T{ 5 1 10 within -> true }T
T{ 10 1 10 within -> false }T
T{ 10 1 10 between -> true }T

\ --- memory ---
variable tv
T{ 42 tv ! tv @ -> 42 }T
T{ 1 tv ! 5 tv +! tv @ -> 6 }T
T{ tv on tv @ -> -1 }T
T{ tv off tv @ -> 0 }T
T{ 5 tv ! tv incr tv @ -> 6 }T
T{ tv decr tv @ -> 5 }T
T{ 3 cells -> 24 }T
T{ 8 cell+ -> 16 }T
T{ 5 chars -> 5 }T
T{ 5 char+ -> 6 }T
create cbuf 3 c, 65 c, 66 c, 67 c,
T{ cbuf c@ -> 3 }T
T{ cbuf 1+ c@ -> 65 }T
create nbuf 11 , 22 ,
T{ nbuf @ -> 11 }T
T{ nbuf cell+ @ -> 22 }T
T{ here 16 allot here swap - -> 16 }T
T{ 7 constant seven seven -> 7 }T
T{ 4 array arr 99 2 arr [] ! 2 arr [] @ -> 99 }T
T{ 3 4 2constant pr pr -> 3 4 }T

\ --- defining words ---
: mk-doubler create , does> @ 2 * ;
21 mk-doubler forty-two
T{ forty-two -> 42 }T

\ --- execution tokens ---
T{ 2 3 ' + execute -> 5 }T
: xt-test ['] * execute ;
T{ 4 5 xt-test -> 20 }T
