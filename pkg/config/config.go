package config

import (
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/garyburd/redigo/redis"
	"github.com/jmoiron/sqlx"
	cfg "zonst/qipai/utils/config"
)

type ErrNotFoundConfig struct {
	Key string
}

func (e ErrNotFoundConfig) Error() string {
	return fmt.Sprintf("not found config:%s", e.Key)
}

type ErrNotPointer struct {
	Key string
}

func (e ErrNotPointer) Error() string {
	return fmt.Sprintf("it's not a pointer")
}

type Config struct {
	DBServers    map[string]cfg.DBServer    `toml:"dbservers"`
	RedisServers map[string]cfg.RedisServer `toml:"redisservers"`
}

// UnmarshalConfig 解析toml配置
func UnmarshalConfig(tomlfile string) (*Config, error) {
	c := &Config{}
	if _, err := toml.DecodeFile(tomlfile, c); err != nil {
		return c, err
	}
	return c, nil
}

// RedisServerConf 获取redis配置
func (c Config) RedisServerConf(key string) (cfg.RedisServer, bool) {
	s, ok := c.RedisServers[key]
	return s, ok
}

// DBServerConf 获取数据库配置
func (c Config) DBServerConf(key string) (cfg.DBServer, bool) {
	s, ok := c.DBServers[key]
	return s, ok
}

func (c Config) Postgres(key string, maxIdleConn int) (*sqlx.DB, error) {
	dbConf, ok := c.DBServerConf(key)
	if !ok {
		return nil, ErrNotFoundConfig{Key: key}
	}
	db, err := dbConf.NewPostgresDB(maxIdleConn)
	if err != nil {
		return nil, fmt.Errorf("连接数据库%v出错, err:%v\n", key, err)
	}
	return db, nil
}

func (c Config) NewRedisPool(key string, maxIdle int) (*redis.Pool, error) {
	conf, ok := c.RedisServers[key]
	if !ok {
		return nil, ErrNotFoundConfig{Key: key}
	}
	pool := conf.NewPool(maxIdle)
	return pool, nil
}
