# snowflake

Twitter snowflake worker - Twitter雪花ID计算器

[twitter announcing](https://blog.twitter.com/engineering/en_us/a/2010/announcing-snowflake.html)


## Introduction

The default Twitter format shown below.

![snowflake.png](snowflake.png)

### Usage

import github.com/github/snowflake

```go

// You can set your epoch time
// but do not set the time before 69 years ago, then you should get the overflowed number
snowflake.SetEpochTime(time.Now()) // default epoch time is 2020-01-01 00:00:00.000

snowflake.SetMaxNode(5, 5, 12) // default 5, 5, 12

worker, _ := snowflake.NewWorker(0, 0)

// get by for
id := worker.Next()

// get by sleep Millisecond
id = worker.NextSleep()
```

### Benchmark

```go
go test -bench=. -benchmem  -run=none


goos: darwin
goarch: amd64
pkg: github.com/go-trellis/snowflake
BenchmarkNext-8                   	 4926423	       244 ns/op	       0 B/op	       0 allocs/op
BenchmarkNextMaxSequence-8        	 9560886	       125 ns/op	       0 B/op	       0 allocs/op
BenchmarkNextNoSequence-8         	    1216	    998952 ns/op	       0 B/op	       0 allocs/op
BenchmarkNextSleep-8              	     925	   1348426 ns/op	       0 B/op	       0 allocs/op
BenchmarkNextSleepMaxSequence-8   	 9538585	       120 ns/op	       0 B/op	       0 allocs/op
BenchmarkNextSleepNoSequence-8    	     872	   1367191 ns/op	       0 B/op	       0 allocs/op
PASS
ok  	github.com/go-trellis/snowflake	8.118s
```