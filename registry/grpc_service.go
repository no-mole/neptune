package registry

import (
	"context"
	"fmt"
	"sync"

	"github.com/no-mole/neptune/logger"
)

// GrpcServiceInstance struct having full info about grpc service instance
type GrpcServiceInstance struct {
	Endpoint string `json:"endpoint"`
}

type ServiceInstanceManager struct {
	sync.RWMutex
	instanceMap map[string]map[string]*GrpcServiceInstance
}

func NewServiceInstanceManager() *ServiceInstanceManager {
	manager := new(ServiceInstanceManager)
	manager.instanceMap = map[string]map[string]*GrpcServiceInstance{}
	return manager
}

func (manager *ServiceInstanceManager) AddNode(key string, instance *GrpcServiceInstance) {
	if instance == nil {
		return
	}
	manager.Lock()
	defer manager.Unlock()
	if _, exist := manager.instanceMap[key]; !exist {
		manager.instanceMap[key] = map[string]*GrpcServiceInstance{}
	}
	manager.instanceMap[key][instance.Endpoint] = instance
}

func (manager *ServiceInstanceManager) DelNode(key string) {
	manager.Lock()
	defer manager.Unlock()
	if _, exist := manager.instanceMap[key]; exist {
		delete(manager.instanceMap, key)
	}
}

func (manager *ServiceInstanceManager) Pick(key string) []*GrpcServiceInstance {
	manager.Lock()
	defer manager.Unlock()
	instances := make([]*GrpcServiceInstance, 0)
	for _, v := range manager.instanceMap[key] {
		instances = append(instances, v)
	}
	return instances
}

func (manager *ServiceInstanceManager) Print() {
	for k, v := range manager.instanceMap {
		for endpoint, obj := range v {
			logger.Info(
				context.Background(),
				"ServiceInstanceManager",
				logger.WithField("msg", fmt.Sprintf("[instance] prefix key:%s endpoint:%s instance:%+v", k, endpoint, obj)),
			)
		}
	}
}
