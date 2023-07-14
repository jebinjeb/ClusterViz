package configurations

import (
	"github.com/kelseyhightower/envconfig"
	gerrors "ClusterViz/internal/pkg/gerror"
	mappers "ClusterViz/internal/pkg/mapper"
)


type ServiceConfigurations struct {
	LogLevel           string `envconfig:"LOG_LEVEL" default:"info"`
	Port               string `envconfig:"PORT" default:"9090"`
	HeaderReadTimeout int
}

func GetServiceConfigurations() (serviceConf *ServiceConfigurations, err error) {
	serviceConf = &ServiceConfigurations{}
	if err = envconfig.Process("", serviceConf); err != nil {
		return nil, gerrors.NewFromError(gerrors.ServiceSetup, err)
	}
	return
}
