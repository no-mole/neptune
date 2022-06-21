package config

import (
	"github.com/no-mole/neptune/config/center"
)

var globalCenterClient center.Client

func initCenterClient() error {
	conf := GetCenterConf()
	if conf.Type == "" {
		return nil
	}

	var err error
	globalCenterClient, err = center.GetClient(conf.Type)
	if err != nil {
		return err
	}

	err = globalCenterClient.Init(
		center.WithNamespace(GlobalConfig.Namespace),
		center.WithEndpoint(conf.Endpoint),
		center.WithAuth(center.Auth{
			Username: conf.UserName,
			Password: conf.Password,
		}),
	)
	return err
}

// GetClient return global config center client
func GetClient() center.Client {
	return globalCenterClient
}
