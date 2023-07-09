To run fibonnaci benchmark:

1. Build: `go build -o ./benchmark/fibonacci ./benchmark`
2. Run Interpretered Version: `./benchmark/fibonacci -engine=eval`
3. Run Compiled Version: `./benchmark/fibonacci -engine=vm`

Same results

```
>> ./fibonacci -engine=vm
>> engine=vm, result=9227465, duration=3.971662833s

>> ./fibonacci -engine=eval
engine=eval, result=9227465, duration=12.701219791s

# 3.2x faster
```
