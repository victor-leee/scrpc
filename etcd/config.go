package etcd

import (
	"context"
	config_backend "github.com/victor-leee/scrpc/github.com/victor-leee/config-backend"
	"gopkg.in/yaml.v2"
	"os"
	"sync"
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
		if err := initServiceConfig(".scrpc.yml"); err != nil {
			return nil, err
		}
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
var serviceConfigInitOnce sync.Once

func initServiceConfig(scrpcFile string) error {
	var err error
	serviceConfigInitOnce.Do(func() {
		var file *os.File
		file, err = os.Open(scrpcFile)
		if err != nil {
			return
		}
		var cfg *rpcConfig
		if err = yaml.NewDecoder(file).Decode(&cfg); err != nil {
			return
		}
		serviceConfig = &defaultImpl{
			rpcCfg:        cfg,
			configService: &config_backend.ConfigBackendServiceImpl{},
		}
	})

	return err
}

func GetConfigClient() ServiceConfig {
	return serviceConfig
}
