# rafiki

Going through Thorsten Ball's Go Interpreter &amp; Compiler Books. The language he designs is called Monkey, and so I named my fork of it Rafiki, after the wise royal monkey advisor in The Lion King

## Overview

### Variable Binding

```
let age = 1;
let name = "Monkey";
let result = 10 \* (20 / 2);
```

### Arrays and Hashes

```
let myArray = [1, 2, 3, 4, 5];
let tob = {"name": "tyler", "age": 28};

myArray[0]       // => 1
tob["name"] // => "Tyler"
```

### Functions

#### Simple Functions

```
// Standard function statement, with return keyword
let add = fn(a, b) { return a + b; };

// But Rafiki also supports implicit return values
let add = fn(a, b) { a + b; };

// Calling functions works just like you think it would
add(1, 2);

```

#### Complex Functions

```
// Recursive functions work. Below also shows off implicit returns in a realistic environment
let fibonacci = fn(x) {
  if (x == 0) {
    0
  } else {
    if (x == 1) {
      1
    } else {
      fibonacci(x - 1) + fibonacci(x - 2);
    }
  }
};

// Higher order functions are also supported! Functions are first-class objects
let twice = fn(f, x) {
  return f(f(x));
};

let addTwo = fn(x) {
  return x + 2;
};

twice(addTwo, 2); // => 6
```
