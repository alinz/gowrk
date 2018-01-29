# gowrk

gowrk is a simple benchmark utility for testing webpages that redirect. Redirect doesn not work with Apache Benchmark.

```
go run main.go --url https://google.com -request 10 -concurrent 2
```

result:

```
Concurrent:              2
Request:                 10
URL:                     https://google.com
------------------
Total time:              1.486381105s
Min Duration:            166.024501ms
Max Duration:            798.886296ms
Average Duration:        296.633581ms
Average Size:            11264 bytes
Errors:                  0
```
