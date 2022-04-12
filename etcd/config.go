package etcd

import (
	"context"
	"errors"
	config_backend "github.com/victor-leee/scrpc/github.com/victor-leee/config-backend"
	"gopkg.in/yaml.v2"
	"os"
)

type ServiceConfig interface {
	Get(ctx context.Context, key string) (*config_backend.GetConfigResponse, error)
}

type defaultImpl struct {
	rpcCfg        *rpcConfig
	configService config_backend.ConfigBackendService
}

func (d *defaultImpl) Get(ctx context.Context, key string) (*config_backend.GetConfigResponse, error) {
	if d.rpcCfg == nil {
		return nil, errors.New("make sure to call Init() before Get()")
	}
	getCfgReq := &config_backend.GetConfigRequest{
		ServiceId:  d.rpcCfg.Service,
		ServiceKey: d.rpcCfg.ServiceKey,
		Key:        key,
	}

	return d.configService.GetConfig(ctx, getCfgReq)
}

type rpcConfig struct {
	Service    string `yaml:"service"`
	ServiceKey string `yaml:"serviceKey"`
}

var serviceConfig ServiceConfig

func Init(scrpcFile string) error {
	file, err := os.Open(scrpcFile)
	if err != nil {
		return err
	}
	var cfg *rpcConfig
	if err = yaml.NewDecoder(file).Decode(&cfg); err != nil {
		return err
	}
	serviceConfig = &defaultImpl{
		rpcCfg:        cfg,
		configService: &config_backend.ConfigBackendServiceImpl{},
	}

	return nil
}

func GetConfigClient() ServiceConfig {
	return serviceConfig
}
