package s3

type Config interface {
	GetAccessKeyId() string
	GetSecretAccessKey() string
}

type Logger interface {
	Debug(args ...interface{})
	Info(args ...interface{})
	Warn(args ...interface{})
	Error(args ...interface{})
}

type Client struct {
	config Config
	logger Logger
}

// New is a s3 client constructor.
func New(config Config, logger Logger) (*Client, error) {
	return &Client{
		config: config,
		logger: logger,
	}, nil
}

func (r *Client) Resize() ([]byte, error) {
	return []byte{}, nil
}
