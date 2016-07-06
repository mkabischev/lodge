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

## Running
```
lodge [-bind=0.0.0.0:20000 [-gc_period=10 -users=/path/to/httpasswd/file]]
```
gc_period - interval for removing expired items. They aren`t available via API since expiration but real removal happens every gc_period.
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

