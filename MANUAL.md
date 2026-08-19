# Forthling-Claude word reference

Every word in `WORDS`, with its stack diagram. 255 words.

## How to read a stack diagram

```
( before -- after )
```

The **top of the stack is on the right**, so `( a b -- b a )` takes `b` off the
top and puts it underneath `a`. Words that also touch the floating-point stack
show it separately:

```
f+     ( F: r1 r2 -- r3 )        only the float stack moves
s>f    ( n -- ) ( F: -- r )      an integer comes off, a float goes on
f<     ( -- flag ) ( F: r1 r2 -- )   floats in, an integer flag out
```

Names used in diagrams:

| Name | Meaning |
|---|---|
| `n`, `n1`, `u` | integer cell (unsigned where `u`) |
| `flag` | `0` for false, `-1` for true |
| `addr`, `a-addr` | address in data space |
| `c-addr` | address of a counted string (length byte first) |
| `c` | character (a byte) |
| `r` | floating-point number, on the float stack |
| `xt` | execution token, from `'` or `[']` |
| `"name"` | the word **parses** the next token from the input |

Markers: **I** = immediate (runs while compiling), **C** = compile-only (valid
only inside a definition or an interpreted control structure), **P** = parses
text from the input stream.

## Contents

- [Stack](#stack)
- [Return stack](#return-stack)
- [Arithmetic and logic](#arithmetic-and-logic)
- [Comparison](#comparison)
- [Memory and data space](#memory-and-data-space)
- [Constants](#constants)
- [Floating point](#floating-point)
- [Output](#output)
- [Input and parsing](#input-and-parsing)
- [Terminal](#terminal)
- [Time and randomness](#time-and-randomness)
- [System](#system)
- [Defining words](#defining-words)
- [Compiling and execution tokens](#compiling-and-execution-tokens)
- [Control structures](#control-structures)
- [Strings](#strings)
- [Loading source](#loading-source)
- [Arrays](#arrays)
- [Screen decoration](#screen-decoration)

## Stack

| Word | Stack | Description |
|---|---|---|
| `dup` | `( n -- n n )` | copy the top item |
| `?dup` | `( n -- n n )` or `( 0 -- 0 )` | copy the top item only if it is non-zero |
| `drop` | `( n -- )` | discard the top item |
| `swap` | `( n1 n2 -- n2 n1 )` | exchange the top two |
| `over` | `( n1 n2 -- n1 n2 n1 )` | copy the second item to the top |
| `nip` | `( n1 n2 -- n2 )` | discard the second item |
| `tuck` | `( n1 n2 -- n2 n1 n2 )` | copy the top item under the second |
| `rot` | `( n1 n2 n3 -- n2 n3 n1 )` | rotate the third item to the top |
| `-rot` | `( n1 n2 n3 -- n3 n1 n2 )` | rotate the top item down to third |
| `2dup` | `( n1 n2 -- n1 n2 n1 n2 )` | copy the top pair |
| `2drop` | `( n1 n2 -- )` | discard the top pair |
| `2swap` | `( n1 n2 n3 n4 -- n3 n4 n1 n2 )` | exchange the top two pairs |
| `2over` | `( n1 n2 n3 n4 -- n1 n2 n3 n4 n1 n2 )` | copy the second pair to the top |
| `3dup` | `( n1 n2 n3 -- n1 n2 n3 n1 n2 n3 )` | copy the top three |
| `pick` | `( ... u -- ... n )` | copy the item `u` deep; `0 pick` is `dup` |
| `roll` | `( ... u -- ... n )` | move the item `u` deep to the top; `2 roll` is `rot` |
| `depth` | `( -- n )` | number of items on the data stack |
| `depth!` | `( ... -- )` | empty the data stack |

## Return stack

Values you push here must be removed before the word returns, and `>r` / `r>`
must not straddle the body of a `do ... loop` (the loop keeps its counters on
the same stack).

| Word | Stack | Description |
|---|---|---|
| `>r` | `( n -- ) ( R: -- n )` | move the top item to the return stack |
| `r>` | `( -- n ) ( R: n -- )` | move it back |
| `r@` | `( -- n ) ( R: n -- n )` | copy the top of the return stack |
| `2>r` | `( n1 n2 -- ) ( R: -- n1 n2 )` | move a pair to the return stack |
| `2r>` | `( -- n1 n2 ) ( R: n1 n2 -- )` | move a pair back |
| `rdepth` | `( -- n )` | number of items on the return stack |

## Arithmetic and logic

Division is **floored**: `-7 2 /` is `-4` and `-7 2 mod` is `1`.

| Word | Stack | Description |
|---|---|---|
| `+` | `( n1 n2 -- n3 )` | add |
| `-` | `( n1 n2 -- n3 )` | subtract `n2` from `n1` |
| `*` | `( n1 n2 -- n3 )` | multiply |
| `/` | `( n1 n2 -- n3 )` | floored divide; errors on zero |
| `mod` | `( n1 n2 -- n3 )` | floored remainder; errors on zero |
| `/mod` | `( n1 n2 -- rem quot )` | remainder and quotient together |
| `*/` | `( n1 n2 n3 -- n4 )` | `n1 * n2 / n3` with a 64-bit intermediate |
| `negate` | `( n -- -n )` | change sign |
| `abs` | `( n -- u )` | absolute value |
| `min` | `( n1 n2 -- n3 )` | smaller of two |
| `max` | `( n1 n2 -- n3 )` | larger of two |
| `1+` | `( n -- n+1 )` | increment |
| `1-` | `( n -- n-1 )` | decrement |
| `2*` | `( n -- n*2 )` | shift left one bit |
| `2/` | `( n -- n/2 )` | arithmetic shift right one bit |
| `lshift` | `( n u -- n2 )` | shift left `u` bits |
| `rshift` | `( n u -- n2 )` | logical shift right `u` bits |
| `and` | `( n1 n2 -- n3 )` | bitwise and |
| `or` | `( n1 n2 -- n3 )` | bitwise or |
| `xor` | `( n1 n2 -- n3 )` | bitwise exclusive or |
| `invert` | `( n -- n2 )` | bitwise complement |
| `sq` | `( n -- n*n )` | square |
| `0max` | `( n -- n2 )` | clamp negatives to zero |

## Comparison

All of these return `-1` for true and `0` for false.

| Word | Stack | Description |
|---|---|---|
| `=` | `( n1 n2 -- flag )` | equal |
| `<>` | `( n1 n2 -- flag )` | not equal |
| `<` | `( n1 n2 -- flag )` | less than |
| `>` | `( n1 n2 -- flag )` | greater than |
| `<=` | `( n1 n2 -- flag )` | less than or equal |
| `>=` | `( n1 n2 -- flag )` | greater than or equal |
| `u<` | `( u1 u2 -- flag )` | unsigned less than |
| `u>` | `( u1 u2 -- flag )` | unsigned greater than |
| `0=` | `( n -- flag )` | equal to zero |
| `0<>` | `( n -- flag )` | not zero |
| `0<` | `( n -- flag )` | negative |
| `0>` | `( n -- flag )` | positive |
| `0<=` | `( n -- flag )` | zero or negative |
| `0>=` | `( n -- flag )` | zero or positive |
| `within` | `( n lo hi -- flag )` | `lo <= n < hi`, upper bound excluded |
| `between` | `( n lo hi -- flag )` | `lo <= n <= hi`, both bounds included |

## Memory and data space

Data space is byte-addressed, one cell is 8 bytes, and cells are stored
little-endian on every platform.

| Word | Stack | Description |
|---|---|---|
| `@` | `( addr -- n )` | fetch a cell |
| `!` | `( n addr -- )` | store a cell |
| `c@` | `( addr -- c )` | fetch a byte |
| `c!` | `( c addr -- )` | store a byte |
| `+!` | `( n addr -- )` | add `n` to the cell at `addr` |
| `f@` | `( addr -- ) ( F: -- r )` | fetch a float from a cell |
| `f!` | `( addr -- ) ( F: r -- )` | store a float into a cell |
| `?` | `( addr -- )` | print the cell at `addr` |
| `on` | `( addr -- )` | store `-1` |
| `off` | `( addr -- )` | store `0` |
| `incr` | `( addr -- )` | add one to the cell |
| `decr` | `( addr -- )` | subtract one from the cell |
| `here` | `( -- addr )` | next free address in data space |
| `allot` | `( n -- )` | reserve `n` bytes; a negative `n` releases them |
| `,` | `( n -- )` | align, then compile a cell into data space |
| `c,` | `( c -- )` | compile a byte |
| `f,` | `( F: r -- )` | align, then compile a float |
| `align` | `( -- )` | round `here` up to a cell boundary |
| `cells` | `( n -- n*8 )` | convert a cell count to bytes |
| `cell+` | `( addr -- addr+8 )` | next cell |
| `cell-` | `( addr -- addr-8 )` | previous cell |
| `chars` | `( n -- n )` | convert a character count to bytes (a no-op here) |
| `char+` | `( addr -- addr+1 )` | next character |
| `floats` | `( n -- n*8 )` | convert a float count to bytes |
| `float+` | `( addr -- addr+8 )` | next float |
| `move` | `( src dst n -- )` | copy `n` bytes, overlap-safe |
| `fill` | `( addr n c -- )` | set `n` bytes to `c` |
| `erase` | `( addr n -- )` | set `n` bytes to zero |
| `count` | `( c-addr -- addr len )` | split a counted string into address and length |
| `bounds` | `( addr len -- limit start )` | turn address and length into `do ... loop` limits |
| `pad` | `( -- addr )` | 256-byte scratch buffer; `word` and `accept` use it too |
| `state` | `( -- addr )` | cell holding compilation state: `0` interpreting |
| `base` | `( -- addr )` | cell holding the current number base |
| `decimal` | `( -- )` | set `base` to 10 |
| `hex` | `( -- )` | set `base` to 16 |

## Constants

| Word | Stack | Value |
|---|---|---|
| `true` | `( -- -1 )` | all bits set |
| `false` | `( -- 0 )` | zero |
| `bl` | `( -- 32 )` | space |
| `cell` | `( -- 8 )` | bytes per cell |
| `bel` | `( -- 7 )` | bell |
| `bs` | `( -- 8 )` | backspace |
| `tab-char` | `( -- 9 )` | tab |
| `nl` | `( -- 10 )` | newline |
| `esc` | `( -- 27 )` | escape |

## Floating point

Floats live on their own stack. A number literal containing `.`, `e` or `E`
that parses as a float goes there, so `1.5`, `-2.0e3` and `1e-4` are floats
while `15` is an integer. Comparisons consume floats and leave an ordinary flag
on the data stack.

| Word | Stack | Description |
|---|---|---|
| `f+` | `( F: r1 r2 -- r3 )` | add |
| `f-` | `( F: r1 r2 -- r3 )` | subtract `r2` from `r1` |
| `f*` | `( F: r1 r2 -- r3 )` | multiply |
| `f/` | `( F: r1 r2 -- r3 )` | divide; errors on zero |
| `fmod` | `( F: r1 r2 -- r3 )` | floating-point remainder |
| `fmin` | `( F: r1 r2 -- r3 )` | smaller of two |
| `fmax` | `( F: r1 r2 -- r3 )` | larger of two |
| `f**` | `( F: r1 r2 -- r3 )` | `r1` raised to the power `r2` |
| `fatan2` | `( F: y x -- angle )` | angle of the point `x,y` in radians |
| `fnegate` | `( F: r -- -r )` | change sign |
| `fabs` | `( F: r -- \|r\| )` | absolute value |
| `fsqrt` | `( F: r -- r2 )` | square root |
| `fsin` | `( F: r -- r2 )` | sine, radians |
| `fcos` | `( F: r -- r2 )` | cosine, radians |
| `ftan` | `( F: r -- r2 )` | tangent, radians |
| `fatan` | `( F: r -- r2 )` | arctangent, radians |
| `fexp` | `( F: r -- r2 )` | e to the power `r` |
| `fln` | `( F: r -- r2 )` | natural log; errors on zero or negative |
| `flog` | `( F: r -- r2 )` | log base 10; errors on zero or negative |
| `floor` | `( F: r -- r2 )` | round down to a whole number |
| `fround` | `( F: r -- r2 )` | round to nearest, halves away from zero |
| `f2*` | `( F: r -- r*2 )` | double |
| `f2/` | `( F: r -- r/2 )` | halve |
| `fsq` | `( F: r -- r*r )` | square |
| `deg>rad` | `( F: deg -- rad )` | convert degrees to radians |
| `f<` | `( -- flag ) ( F: r1 r2 -- )` | less than |
| `f>` | `( -- flag ) ( F: r1 r2 -- )` | greater than |
| `f=` | `( -- flag ) ( F: r1 r2 -- )` | equal |
| `f<=` | `( -- flag ) ( F: r1 r2 -- )` | less than or equal |
| `f>=` | `( -- flag ) ( F: r1 r2 -- )` | greater than or equal |
| `f0<` | `( -- flag ) ( F: r -- )` | negative |
| `f0=` | `( -- flag ) ( F: r -- )` | zero |
| `fdup` | `( F: r -- r r )` | copy the top float |
| `fdrop` | `( F: r -- )` | discard the top float |
| `fswap` | `( F: r1 r2 -- r2 r1 )` | exchange the top two |
| `fover` | `( F: r1 r2 -- r1 r2 r1 )` | copy the second to the top |
| `fnip` | `( F: r1 r2 -- r2 )` | discard the second |
| `frot` | `( F: r1 r2 r3 -- r2 r3 r1 )` | rotate the third to the top |
| `fdepth` | `( -- n )` | number of items on the float stack |
| `s>f` | `( n -- ) ( F: -- r )` | integer to float |
| `f>s` | `( -- n ) ( F: r -- )` | float to integer, truncating toward zero |
| `f.` | `( F: r -- )` | print a float, then a space |
| `fe.` | `( F: r -- )` | print in scientific notation, then a space |
| `f.r` | `( width places -- ) ( F: r -- )` | print right-aligned in `width` with `places` decimals |
| `f.s` | `( -- )` | show the float stack without changing it |
| `pi` | `( F: -- 3.14159… )` | pi |
| `e` | `( F: -- 2.71828… )` | Euler's number |

## Output

Output is buffered. Input words and `ms` flush it for you; call `flush` if your
loop does neither.

| Word | Stack | Description |
|---|---|---|
| `.` | `( n -- )` | print a number in the current base, then a space |
| `u.` | `( u -- )` | print as unsigned |
| `.r` | `( n width -- )` | print right-aligned in `width` columns, no trailing space |
| `emit` | `( c -- )` | print one character |
| `cr` | `( -- )` | newline |
| `space` | `( -- )` | one space |
| `spaces` | `( n -- )` | `n` spaces |
| `type` | `( addr len -- )` | print `len` bytes from `addr` |
| `.s` | `( -- )` | show the data stack without changing it |
| `flush` | `( -- )` | push buffered output to the terminal now |

## Input and parsing

| Word | Stack | Description |
|---|---|---|
| `key` | `( -- c )` | wait for one character; `-1` at end of input |
| `key?` | `( -- flag )` | true if a character is waiting; never blocks |
| `accept` | `( addr max -- len )` | read a line into `addr`, at most `max` bytes; `0` at end of input |
| `word` | `( c -- c-addr )` | **P** parse the next token from the input up to delimiter `c`, leave a counted string at `pad` |
| `number` | `( addr len -- n flag )` | convert a string in the current base; `flag` is false if it is not a number |
| `char` | `( "name" -- c )` | **P** first character of the next token |
| `[char]` | `( -- c )` | **I P** the same, compiled as a literal |

## Terminal

| Word | Stack | Description |
|---|---|---|
| `page` | `( -- )` | clear the screen, cursor home |
| `at-xy` | `( x y -- )` | move the cursor, `0 0` is top left |
| `term-size` | `( -- cols rows )` | window size; falls back to `$COLUMNS`/`$LINES`, then 80x24 |
| `raw-on` | `( -- )` | character-at-a-time input, no echo; silently ignored when not a terminal |
| `raw-off` | `( -- )` | restore normal line input |
| `cursor-on` | `( -- )` | show the cursor |
| `cursor-off` | `( -- )` | hide the cursor |

## Time and randomness

| Word | Stack | Description |
|---|---|---|
| `ms` | `( n -- )` | flush output and sleep `n` milliseconds |
| `ticks` | `( -- n )` | milliseconds since the program started |
| `random` | `( n -- u )` | pseudo-random number in `0 .. n-1`; errors if `n` is not positive |
| `seed` | `( n -- )` | set the generator seed, for repeatable runs |
| `randomize` | `( -- )` | seed from the clock |

## System

| Word | Stack | Description |
|---|---|---|
| `bye` | `( -- )` | leave the program with status 0 |
| `quit` | `( -- )` | same as `bye` here; it does **not** return to the interpreter as in standard Forth |
| `abort` | `( -- )` | raise an error: clears the stacks, returns to the prompt, fails a script |
| `words` | `( -- )` | list the dictionary |
| `argc` | `( -- n )` | number of command line arguments after the script name |
| `arg` | `( n -- addr len )` | argument `n`, counting from 0; an empty string if out of range |

## Defining words

| Word | Stack | Description |
|---|---|---|
| `:` | `( "name" -- )` | **I P** begin a definition and start compiling |
| `;` | `( -- )` | **I C** finish the definition |
| `immediate` | `( -- )` | mark the last definition immediate |
| `create` | `( "name" -- )` | **P** make a word that pushes the address of its data space |
| `variable` | `( "name" -- )` | **P** create a one-cell variable, initialised to zero; the word returns its address |
| `fvariable` | `( "name" -- )` | **P** the same, intended for use with `f@` and `f!` |
| `constant` | `( n "name" -- )` | **P** create a word that pushes `n` |
| `fconstant` | `( "name" -- ) ( F: r -- )` | **P** create a word that pushes `r` onto the float stack |
| `2constant` | `( n1 n2 "name" -- )` | **P** create a word that pushes both, `n1` first |
| `does>` | `( -- )` | **I C** give the most recently created word the run-time behaviour that follows |

`does>` is how you build your own defining words. The child word gets its data
address pushed before the `does>` code runs:

```forth
: mk-doubler ( n "name" -- ) create , does> @ 2 * ;
21 mk-doubler forty-two
forty-two .            \ prints 42
```

## Compiling and execution tokens

| Word | Stack | Description |
|---|---|---|
| `'` | `( "name" -- xt )` | **P** find a word and return its execution token |
| `[']` | `( -- xt )` | **I C P** the same, compiled as a literal |
| `execute` | `( xt -- )` | run the word |
| `compile,` | `( xt -- )` | append the word to the definition being compiled |
| `postpone` | `( "name" -- )` | **I C P** compile the word's compilation behaviour |
| `literal` | `( n -- )` | **I C** compile the value on the stack as a literal |
| `fliteral` | `( -- ) ( F: r -- )` | **I C** compile a float literal |
| `recurse` | `( -- )` | **I C** call the definition currently being compiled |
| `[` | `( -- )` | **I** stop compiling, start interpreting |
| `]` | `( -- )` | start compiling |
| `(` | `( "ccc)" -- )` | **I P** comment up to the closing bracket |
| `\` | `( "ccc" -- )` | **I P** comment to the end of the line |

## Control structures

These are compiling words. At the prompt they still work: the line is compiled
into a scratch definition and run when the structure closes.

| Word | Stack | Description |
|---|---|---|
| `if` | `( flag -- )` | **I C** run the following code when `flag` is true |
| `else` | `( -- )` | **I C** alternative branch |
| `then` | `( -- )` | **I C** end of an `if` |
| `begin` | `( -- )` | **I C** start of an indefinite loop |
| `until` | `( flag -- )` | **I C** loop back to `begin` while `flag` is false |
| `again` | `( -- )` | **I C** loop back to `begin` forever |
| `while` | `( flag -- )` | **I C** leave the loop when `flag` is false |
| `repeat` | `( -- )` | **I C** loop back to `begin`, closing a `while` |
| `do` | `( limit start -- )` | **I C** counted loop; the body always runs at least once |
| `?do` | `( limit start -- )` | **I C** counted loop, skipped entirely when `limit = start` |
| `loop` | `( -- )` | **I C** add one to the index and repeat until it reaches the limit |
| `+loop` | `( n -- )` | **I C** add `n` to the index; a negative `n` counts down |
| `i` | `( -- n )` | **C** index of the innermost loop |
| `j` | `( -- n )` | **C** index of the next loop out |
| `leave` | `( -- )` | **I C** exit the loop immediately |
| `unloop` | `( -- )` | **C** discard the loop counters, needed before `exit` inside a loop |
| `exit` | `( -- )` | **C** return from the current definition |
| `case` | `( -- )` | **I C** start a selector |
| `of` | `( x1 x2 -- )` matched, `( x1 x2 -- x1 )` not | **I C** run this arm when the two values match, dropping the selector |
| `endof` | `( -- )` | **I C** end of an arm |
| `endcase` | `( n -- )` | **I C** end of the selector, dropping the value |

```forth
: name ( n -- )
  case
    1 of ." one"   endof
    2 of ." two"   endof
    ." many"                 \ default arm; endcase drops the selector
  endcase ;
```

## Strings

| Word | Stack | Description |
|---|---|---|
| `s"` | `( "ccc\"" -- addr len )` | **I P** an address and length pair; permanent inside a definition, in a rotating buffer when interpreted |
| `c"` | `( "ccc\"" -- c-addr )` | **I P** a counted string |
| `."` | `( "ccc\"" -- )` | **I P** print the text |
| `.(` | `( "ccc)" -- )` | **I P** print the text right now, even while compiling |
| `abort"` | `( flag -- )` | **I C P** raise an error with this message when `flag` is true |
| `str=` | `( a1 n1 a2 n2 -- flag )` | compare two strings for equality |

## Loading source

| Word | Stack | Description |
|---|---|---|
| `include` | `( "filename" -- )` | **P** read and interpret a file |
| `included` | `( addr len -- )` | the same, with the name as a string |
| `evaluate` | `( addr len -- )` | interpret a string as Forth source |

## Arrays

Built on `create`, so they cost nothing at run time beyond the address maths.

| Word | Stack | Description |
|---|---|---|
| `array` | `( n "name" -- )` | **P** create an array of `n` cells |
| `[]` | `( i addr -- addr' )` | address of cell `i` in an array |
| `barray` | `( n "name" -- )` | **P** create an array of `n` bytes |
| `b[]` | `( i addr -- addr' )` | address of byte `i` |

```forth
10 array counts
5 3 counts [] !          \ counts[3] := 5
3 counts [] @ .          \ prints 5
```

## Screen decoration

Colours are 256-colour palette indices, the usual `0-15` basics, `16-231`
colour cube, `232-255` greyscale.

| Word | Stack | Description |
|---|---|---|
| `fg` | `( n -- )` | set the foreground colour |
| `bg` | `( n -- )` | set the background colour |
| `normal` | `( -- )` | reset all attributes |
| `bright` | `( -- )` | bold or bright attribute |
| `esc[` | `( -- )` | emit the escape and bracket that start an ANSI sequence |
| `tab` | `( -- )` | emit a tab |
| `bell` | `( -- )` | ring the terminal bell |
| `star` | `( -- )` | emit one `*` |
| `stars` | `( n -- )` | emit `n` stars |
| `dots` | `( n -- )` | emit `n` full stops |

## Errors

Any of these raise an error, which prints a message, empties the data, float
and return stacks, abandons a half-finished definition, and returns to the
prompt (or exits a script with status 1):

- `undefined word: xyz`
- `data stack underflow`, `data stack overflow`, and the return and float
  equivalents
- `division by zero`, `float division by zero`
- `address out of range: n`
- the message given to `abort"`, or `aborted` from `abort`
- compiler complaints such as `unclosed control structure at ;` or
  `then without matching opener`
