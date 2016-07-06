package client

var (
	operationAuth    = "AUTH"
	operationSet     = "SET"
	operationGet     = "GET"
	operationHSet    = "HSET"
	operationHGet    = "HGET"
	operationHGetAll = "HGETALL"
	operationKeys    = "KEYS"
	operationDelete  = "DELETE"
)

// Config is a struct representing configuration for logde client
type Config struct {
	Addr           string
	MaxConnections uint
	Username       string
	Password       string
}

func DefaultConfig() Config {
	return Config{
		Addr:           "localhost:20000",
		MaxConnections: 10,
	}
}

// Client is client for logde server. Client uses connection pooling, so it can be safety used in multiply goroutines.
type Client struct {
	pool     *pool
	username string
	password string
}

// New constructs new Client with specified configuration
func New(config Config) *Client {
	return &Client{
		pool:     newPool(config.Addr, 10),
		username: config.Username,
		password: config.Password,
	}
}

// Get method
func (c *Client) Get(key string) (interface{}, error) {
	result, err := c.call(operationGet, args(key), nil)
	if err != nil {
		return nil, err
	}

	return result[0], nil
}

// Set method
func (c *Client) Set(key, value string, ttl int64) error {
	_, err := c.call(operationSet, args(key, ttl, len(value)), value)

	return err
}

// Keys method
func (c *Client) Keys() ([]string, error) {
	return c.call(operationKeys, nil, nil)
}

// HSet method
func (c *Client) HSet(key, field, value string) error {
	_, err := c.call(operationHSet, args(key, field, len(value)), value)

	return err
}

// HGet method
func (c *Client) HGet(key, field string) (interface{}, error) {
	result, err := c.call(operationHGet, args(key, field), nil)
	if err != nil {
		return nil, err
	}

	return result[0], nil
}

func (c *Client) HGetAll(key string) (map[string]interface{}, error) {
	result, err := c.call(operationHGetAll, args(key), nil)
	if err != nil {
		return nil, err
	}

	hash := make(map[string]interface{}, len(result)/2)
	for i := 0; i < len(result); i = i + 2 {
		hash[result[i]] = result[i+1]
	}

	return hash, nil
}

// Delete method
func (c *Client) Delete(key string) error {
	_, err := c.call(operationDelete, args(key), nil)

	return err
}

// call executes command on server. If data is passed then it is added after request:
// COMMAND_NAME arg1 arg2 arg2\r\n
// data\r\n
func (c *Client) call(operation string, arguments []interface{}, data interface{}) ([]string, error) {
	conn, isNew, err := c.pool.get()
	if err != nil {
		return nil, err
	}

	proto := newConnection(conn)

	// check is authentication is required
	if isNew && c.username != "" {
		if _, err := proto.send(operationAuth, args(c.username, c.password), nil); err != nil {
			return nil, err
		}
	}

	result, err := proto.send(operation, arguments, data)
	if err != nil {
		return nil, err
	}
	c.pool.put(conn)
	return result, nil
}

// args is tiny helper that adds some syntax-sugar :)
func args(a ...interface{}) []interface{} {
	return a
}
