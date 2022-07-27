package config

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

var ErrConfigRead = errors.New("unable to read config file")

type Config struct {
	Logger LoggerConf
	HTTP   HTTPConf
	Cache  CacheConf
	S3     S3Conf
}

type LoggerConf struct {
	Level   string
	File    string
	Size    int
	Backups int
	Age     int
}

type HTTPConf struct {
	Host string
	Port string
}

type CacheConf struct {
	Capacity int64
	Path     string
}

type S3Conf struct {
	AccessKeyId     string
	SecretAccessKey string
}

func NewConfig(path string) (*Config, error) {
	viper.SetConfigFile(path)

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("%w: %s", ErrConfigRead, path)
	}

	return &Config{
		LoggerConf{
			viper.GetString("logger.level"),
			viper.GetString("logger.file"),
			viper.GetInt("logger.size"),
			viper.GetInt("logger.backups"),
			viper.GetInt("logger.age"),
		},
		HTTPConf{
			viper.GetString("http.host"),
			viper.GetString("http.port"),
		},
		CacheConf{
			viper.GetInt64("cache.capacity"),
			viper.GetString("cache.path"),
		},
		S3Conf{
			viper.GetString("s3.access_key_id"),
			viper.GetString("s3.secret_access_key"),
		},
	}, nil
}

func (c *Config) GetLoggerLevel() string {
	return c.Logger.Level
}

func (c *Config) GetLoggerFile() string {
	return c.Logger.File
}

func (c *Config) GetHTTPHost() string {
	return c.HTTP.Host
}

func (c *Config) GetHTTPPort() string {
	return c.HTTP.Port
}

func (c *Config) GetCacheCapacity() int64 {
	return c.Cache.Capacity
}

func (c *Config) GetCachePath() string {
	return c.Cache.Path
}

func (c *Config) GetAccessKeyId() string {
	return c.S3.AccessKeyId
}

func (c *Config) GetSecretAccessKey() string {
	return c.S3.SecretAccessKey
}
