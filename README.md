Logde Project
---

Lodge is key-value in-memory storage with hashes support.
Logde use simple text protocol very similar to memcached one. You connect connect and send commands using telnet or nc.

Available commands:

| Command | Description                | Example                                  |
|---------|----------------------------|------------------------------------------|
| SET     | Sets value to key          | ```SET key1 0 3\r\foo\r\n```             |
| GET     | Reads value                | ```GET foo```                            |
| HSET    | Sets value to hash         | ```HSET key1 field1 5\r\n hello\r\n```   |
| HGET    | Reads value from hash      | ```HGET key1 field1```                   |
| HGETALL | Reads all values from hash | ```HGETALL key1```                       |
| DELETE  | Deletes key                | ```DELETE key1```                        |
| KEYS    | Returns all available keys | ```KEYS```                               |
| EXPIRE  | Set ttl for key	           | ```EXPIRE foo 100```                     |
| AUTH    | Authenticates user         | ```AUTH username password```             |



Some examples.

Let\`s set value `hello` for key `foo` with ttl `100` seconds.

```
$ printf "SET foo 100 5\r\nhello\r\n" | nc localhost 20000
OK
```
 `5` is value size.


Now let\`s retrieve this value.
```
$ echo "GET foo" | nc localhost 20000
VALUES # marker that response contains values
1 # number of values in response
5 # first value length
hello # first value itself
```

## Building

Logde has no dependencies, so it can be easily build with:
```
go install github.com/mkabischev/lodge
```

## Running tests

```
go test ./...
```

Some tests takes additional time for expires logic checking, you can skip them:
```
go test -test.short ./...
```

### Benchmarks

There are some benchmarks for storages:
```
go test -bench=Storage -benchmem -run=NONE ./...


BenchmarkStorageMemorySet-8      3000000               416 ns/op              40 B/op          2 allocs/op
BenchmarkStorageBucketSet-8      5000000               260 ns/op              45 B/op          2 allocs/op
BenchmarkLStorageRUSet-8         3000000               379 ns/op              45 B/op          2 allocs/op
BenchmarkStorageMemoryGet-8     30000000                70.0 ns/op             7 B/op          1 allocs/op
BenchmarkStorageBucketGet-8     10000000               143 ns/op               7 B/op          1 allocs/op
BenchmarkLStorageRUGet-8        10000000               182 ns/op               7 B/op          1 allocs/op
BenchmarkStorageMemoryCombine-8 10000000               286 ns/op              15 B/op          1 allocs/op
BenchmarkStorageBucketCombine-8 10000000               180 ns/op              15 B/op          1 allocs/op
BenchmarkStorageLRUCombine-8     5000000               280 ns/op              16 B/op          1 allocs/op
```

For set command the fastest is BucketStorage with lru buckets. For get command BucketStorage is 2 time slower then simple storage, but in combine mode (80% gets & 20% sets) fastest is still BucketStorage.

Note: 1 of allocs in each benchmark is converting from i to string, so Set commands use only 1 alloc and get use zero allocs.

## Running
```
lodge [-bind=0.0.0.0:20000 [-buckets=100 [-bucket_size=10000 [-users=/path/to/httpasswd/file]]]
```
buckets - number of buckets
bucket_size - number of elements in each bucket
users - if flag is passed, then for all connections first command must be
```
AUTH username password
```

## Using client
```go
config := client.DefaultConfig()
config.Username = "test"
config.Password = "password"

client := client.New(config)
client.Set("foo", "bar", 5)
val, _ := client.Get("foo")
fmt.Println(val)
```

