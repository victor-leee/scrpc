package scrpc

import (
	"os"
	"strconv"
)

type PoolConfig struct {
	InitSize int
	MaxSize  int
}

type TransportConfig struct {
	Protocol string
	Path     string
	PoolCfg  *PoolConfig
}

type Config struct {
	LocalTransportConfig  *TransportConfig
	RemoteTransportConfig *TransportConfig
}

var cfg *Config

func init() {
	cfg = &Config{
		LocalTransportConfig: &TransportConfig{
			Protocol: env("__SCRPC_LOCAL_TRANSPORT_CONFIG_PROTO", "unix"),
			Path:     env("__SCRPC_LOCAL_TRANSPORT_CONFIG_PATH", "/tmp/sc.sock"),
			PoolCfg: &PoolConfig{
				InitSize: str2Int(env("__SCRPC_LOCAL_TRANSPORT_CONFIG_POOL_INIT_SIZE", "10")),
				MaxSize:  str2Int(env("__SCRPC_LOCAL_TRANSPORT_CONFIG_POOL_MAX_SIZE", "50")),
			},
		},
		RemoteTransportConfig: &TransportConfig{
			Protocol: env("__SCRPC_REMOTE_TRANSPORT_CONFIG_PROTO", "tcp"),
			PoolCfg: &PoolConfig{
				InitSize: str2Int(env("__SCRPC_REMOTE_TRANSPORT_CONFIG_POOL_INIT_SIZE", "10")),
				MaxSize:  str2Int(env("__SCRPC_REMOTE_TRANSPORT_CONFIG_POOL_MAX_SIZE", "50")),
			},
		},
	}
}

func GetConfig() *Config {
	return cfg
}

func env(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}

	return fallback
}

func str2Int(str string) int {
	v, _ := strconv.ParseInt(str, 10, 64)
	return int(v)
}
