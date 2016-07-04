package client

var (
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
	addr           string
	maxConnections uint
}

// Client is client for logde server. Client uses connection pooling, so it can be safety used in multiply goroutines.
type Client struct {
	pool *pool
}

// New constructs new Client with specified configuration
func New(config Config) *Client {
	return &Client{
		pool: newPool(config.addr, 10),
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
func (c *Client) HSet(key, field, value string, ttl int64) error {
	_, err := c.call(operationHSet, args(key, field, ttl, len(value)), value)

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
func (c *Client) call(operation string, args []interface{}, data interface{}) ([]string, error) {
	conn, err := c.pool.get()
	if err != nil {
		return nil, err
	}

	result, err := newConnection(conn).send(operation, args, data)
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
